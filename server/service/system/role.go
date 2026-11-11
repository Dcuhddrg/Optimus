package system

import (
	"context"
	"errors"
	"sort"
	"strconv"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// roleLockTestHooks pause a transaction right after it took its role locks,
// for the assignment/deletion interleaving tests.
var roleLockTestHooks struct {
	afterAssignLock func(userID uint)
	afterDeleteLock func(roleID uint)
}

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrProtectedAdminRole = errors.New("admin role is protected")
	// ErrProtectedRole: admin and the default user role cannot be deleted.
	ErrProtectedRole = errors.New("admin and user roles cannot be deleted")
	// ErrProtectedRoleCode: the admin and user role codes cannot change.
	ErrProtectedRoleCode = errors.New("built-in role codes cannot be changed")
	// ErrRoleInUse: remove the role from its users before deleting it.
	ErrRoleInUse = errors.New("role is still assigned to users; remove it from them first")
)

type RoleService struct{}

func (s *RoleService) ListRoles(ctx context.Context) ([]systemModel.SystemRole, error) {
	var roles []systemModel.SystemRole
	err := global.OPTIMUS_DB.WithContext(ctx).Order("id ASC").Find(&roles).Error
	return roles, err
}

func (s *RoleService) CreateRole(ctx context.Context, actorID uint, req systemRequest.RoleRequest, meta AuditMeta) (systemModel.SystemRole, error) {
	role := systemModel.SystemRole{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Approvable:  req.Approvable != nil && *req.Approvable,
	}
	if role.Code == AdminRoleCode && role.Approvable {
		return systemModel.SystemRole{}, ErrProtectedAdminRole
	}
	created, err := s.createRole(ctx, role)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "role.create", ResourceType: "role", ResourceID: resourceID(created.ID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return created, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, actorID uint, roleID uint, req systemRequest.RoleRequest, meta AuditMeta) (systemModel.SystemRole, error) {
	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	// Built-in roles are found by code: admin by the admin checks, user when
	// binding new LDAP employees and by the delete protection.
	if (role.Code == AdminRoleCode || role.Code == DefaultUserRoleCode) && req.Code != "" && req.Code != role.Code {
		return systemModel.SystemRole{}, ErrProtectedRoleCode
	}
	// admin never carries the approval flag: an administrator who approves
	// holds the approval role as well (design 3.5).
	if role.Code == AdminRoleCode && req.Approvable != nil && *req.Approvable {
		return systemModel.SystemRole{}, ErrProtectedAdminRole
	}
	columns := map[string]any{"name": req.Name, "description": req.Description}
	if req.Code != "" {
		columns["code"] = req.Code
	}
	approvableChanged := req.Approvable != nil && *req.Approvable != role.Approvable
	if req.Approvable != nil {
		columns["approvable"] = *req.Approvable
	}
	if err := global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemRole{}).Where("id = ?", roleID).Updates(columns).Error; err != nil {
		return systemModel.SystemRole{}, err
	}
	if approvableChanged {
		// Third parties' approval roles are ignored when authorizing, so the
		// flag changes their effective roles (design 3.1).
		InvalidateAllAuthState()
	}
	recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "role.update", ResourceType: "role", ResourceID: resourceID(roleID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return s.findRoleByID(ctx, roleID)
}

// DeleteRole deletes an unused, unprotected role with its menu grants and
// Casbin policies, as GVA's DeleteAuthority (design 3.2). Approval-chain
// references are checked once approval chains exist (#4).
func (s *RoleService) DeleteRole(ctx context.Context, actorID uint, roleID uint, meta AuditMeta) error {
	event := failedAudit(actorID, meta, "role.delete", "role", roleID)
	defer func() { recordAudit(ctx, event) }()
	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return err
	}
	if role.Code == AdminRoleCode || role.Code == DefaultUserRoleCode {
		event.Detail = auditDetails("protected_role", nil)
		return ErrProtectedRole
	}
	// The holder check and the delete run in one transaction holding an
	// exclusive lock on the role row. Assignments take a shared lock on the
	// roles they bind (replaceUserRoles), so the two serialize: an assignment
	// committed first makes this delete see a holder; a delete committed
	// first makes the assignment skip the vanished role.
	inUse := false
	err = global.OPTIMUS_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked systemModel.SystemRole
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", roleID).First(&locked).Error; err != nil {
			return err
		}
		if hook := roleLockTestHooks.afterDeleteLock; hook != nil {
			hook(roleID)
		}
		var holders int64
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Model(&systemModel.SystemUserRole{}).Where("role_id = ?", roleID).Count(&holders).Error; err != nil {
			return err
		}
		if holders > 0 {
			inUse = true
			return ErrRoleInUse
		}
		if err := tx.Where("role_id = ?", roleID).Delete(&systemModel.SystemRolePermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("ptype = ? AND v0 = ?", "p", strconv.Itoa(int(roleID))).Delete(&gormadapter.CasbinRule{}).Error; err != nil {
			return err
		}
		return tx.Delete(&systemModel.SystemRole{}, roleID).Error
	})
	if inUse {
		event.Detail = auditDetails("role_in_use", nil)
		return ErrRoleInUse
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return ErrRoleNotFound
	}
	if err != nil {
		event.Detail = auditDetails("delete_failed", nil)
		return err
	}
	// Commit, reload Casbin, invalidate, return (design 3.1).
	reloadErr := CasbinServiceApp.FreshCasbin()
	InvalidateAllAuthState()
	event.Result = "success"
	event.Detail = auditDetails("deleted", nil)
	if reloadErr != nil {
		event.Detail = auditDetails("deleted_reload_pending", nil)
	}
	return reloadErr
}

// ListRolesWithPermissions returns every role with its permission codes, for
// the role editor.
func (s *RoleService) ListRolesWithPermissions(ctx context.Context) ([]systemResponse.RoleListItem, error) {
	roles, err := s.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]systemResponse.RoleListItem, 0, len(roles))
	for _, role := range roles {
		codes, err := s.permissionCodesForRole(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, systemResponse.RoleListItem{SystemRole: role, PermissionCodes: codes})
	}
	return items, nil
}

// ListPermissions returns the whole permission catalog (all modules), as
// stored, for the role editor.
func (s *RoleService) ListPermissions(ctx context.Context) ([]systemModel.SystemPermission, error) {
	var permissions []systemModel.SystemPermission
	err := global.OPTIMUS_DB.WithContext(ctx).Order("sort ASC, id ASC").Find(&permissions).Error
	return permissions, err
}

func (s *RoleService) UpdateRolePermissions(ctx context.Context, actorID uint, roleID uint, permissionCodes []string, meta AuditMeta) error {
	event := failedAudit(actorID, meta, "role.permissions.update", "role", roleID)
	defer func() { recordAudit(ctx, event) }()
	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return err
	}
	if role.Code == AdminRoleCode && !containsAllSystemPermissions(permissionCodes) {
		event.Detail = auditDetails("admin_role_protected", nil)
		return ErrProtectedAdminRole
	}
	permissions, err := s.permissionsByCodes(ctx, permissionCodes)
	if err != nil {
		event.Detail = auditDetails("permission_lookup_failed", nil)
		return err
	}
	before, beforeErr := s.permissionCodesForRole(ctx, roleID)
	err = s.replaceRolePermissions(ctx, roleID, permissions)
	reloadPending := errors.Is(err, ErrCasbinReloadPending)
	if err != nil && !reloadPending {
		event.Detail = auditDetails("update_failed", nil)
		return err
	}
	// Committed (and reloaded unless pending): order is commit, reload Casbin,
	// invalidate, return (design 3.1).
	InvalidateAllAuthState()
	after, afterErr := s.permissionCodesForRole(ctx, roleID)
	event.Result = "success"
	reason := "updated"
	if reloadPending {
		reason = "updated_reload_pending"
	}
	if beforeErr != nil || afterErr != nil {
		event.Detail = auditDetails("snapshot_unavailable", nil)
	} else {
		event.Detail = auditDetails(reason, map[string]auditChange{"permissionCodes": auditSetChange(before, after)})
	}
	return err
}

// containsAllSystemPermissions: the admin role always holds every system.*
// permission so the built-in admin can always recover (design 3.4); other
// modules' permissions stay editable.
func containsAllSystemPermissions(codes []string) bool {
	set := make(map[string]bool, len(codes))
	for _, code := range codes {
		set[code] = true
	}
	for _, required := range SystemPermissionCodes() {
		if !set[required] {
			return false
		}
	}
	return true
}

func (s *RoleService) findRoleByID(ctx context.Context, id uint) (systemModel.SystemRole, error) {
	var role systemModel.SystemRole
	err := global.OPTIMUS_DB.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemRole{}, ErrRoleNotFound
	}
	return role, err
}

func (s *RoleService) createRole(ctx context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error) {
	err := global.OPTIMUS_DB.WithContext(ctx).Create(&role).Error
	return role, err
}

func (s *RoleService) updateRole(ctx context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error) {
	err := global.OPTIMUS_DB.WithContext(ctx).Save(&role).Error
	return role, err
}

func (s *RoleService) permissionsByCodes(ctx context.Context, codes []string) ([]systemModel.SystemPermission, error) {
	var permissions []systemModel.SystemPermission
	err := global.OPTIMUS_DB.WithContext(ctx).Where("code IN ?", codes).Find(&permissions).Error
	return permissions, err
}

// replaceRolePermissions stores menu grants in sys_role_permissions and API
// grants as Casbin policies (GVA UpdateCasbin), then reloads the enforcer.
func (s *RoleService) replaceRolePermissions(ctx context.Context, roleID uint, permissions []systemModel.SystemPermission) error {
	err := global.OPTIMUS_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&systemModel.SystemRolePermission{}).Error; err != nil {
			return err
		}
		for _, permission := range permissions {
			if permission.Type == "api" {
				continue
			}
			if err := tx.Create(&systemModel.SystemRolePermission{RoleID: roleID, PermissionID: permission.ID}).Error; err != nil {
				return err
			}
		}
		return CasbinServiceApp.UpdateCasbin(tx, roleID, APIPolicies(permissions))
	})
	if err != nil {
		return err
	}
	return CasbinServiceApp.FreshCasbin()
}

// permissionCodesForRole returns menu codes from sys_role_permissions plus the
// catalog codes of the role's Casbin API policies.
func (s *RoleService) permissionCodesForRole(ctx context.Context, roleID uint) ([]string, error) {
	db := global.OPTIMUS_DB.WithContext(ctx)
	permissionIDs := db.Model(&systemModel.SystemRolePermission{}).Select("permission_id").Where("role_id = ?", roleID)
	codes := make([]string, 0)
	if err := global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemPermission{}).Where("id IN (?)", permissionIDs).Pluck("code", &codes).Error; err != nil {
		return nil, err
	}
	apiCodes, err := apiPermissionCodesForRole(global.OPTIMUS_DB.WithContext(ctx), roleID)
	if err != nil {
		return nil, err
	}
	codes = append(codes, apiCodes...)
	sort.Strings(codes)
	return codes, nil
}
