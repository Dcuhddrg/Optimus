package system

import (
	"context"
	"errors"
	"sort"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"

	"gorm.io/gorm"
)

var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrProtectedAdminRole = errors.New("admin role is protected")
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
		Status:      DefaultRoleStatus,
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
	if role.Code == "admin" && req.Code != "" && req.Code != "admin" {
		return systemModel.SystemRole{}, ErrProtectedAdminRole
	}
	if req.Code != "" {
		role.Code = req.Code
	}
	role.Name = req.Name
	role.Description = req.Description
	updated, err := s.updateRole(ctx, role)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "role.update", ResourceType: "role", ResourceID: resourceID(roleID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return updated, nil
}

func (s *RoleService) UpdateRoleStatus(ctx context.Context, actorID uint, roleID uint, status string, meta AuditMeta) (systemModel.SystemRole, error) {
	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	if role.Code == "admin" && status != DefaultRoleStatus {
		return systemModel.SystemRole{}, ErrProtectedAdminRole
	}
	role.Status = status
	updated, err := s.updateRole(ctx, role)
	if err != nil {
		return systemModel.SystemRole{}, err
	}
	recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "role.status.update", ResourceType: "role", ResourceID: resourceID(roleID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return updated, nil
}

func (s *RoleService) UpdateRolePermissions(ctx context.Context, actorID uint, roleID uint, permissionCodes []string, meta AuditMeta) error {
	event := failedAudit(actorID, meta, "role.permissions.update", "role", roleID)
	defer func() { recordAudit(ctx, event) }()
	role, err := s.findRoleByID(ctx, roleID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return err
	}
	if role.Code == "admin" && !containsAllCoreAdminPermissions(permissionCodes) {
		event.Detail = auditDetails("admin_role_protected", nil)
		return ErrProtectedAdminRole
	}
	permissions, err := s.permissionsByCodes(ctx, permissionCodes)
	if err != nil {
		event.Detail = auditDetails("permission_lookup_failed", nil)
		return err
	}
	before, beforeErr := s.permissionCodesForRole(ctx, roleID)
	if err := s.replaceRolePermissions(ctx, roleID, permissions); err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return err
	}
	after, afterErr := s.permissionCodesForRole(ctx, roleID)
	event.Result = "success"
	if beforeErr != nil || afterErr != nil {
		event.Detail = auditDetails("snapshot_unavailable", nil)
	} else {
		event.Detail = auditDetails("updated", map[string]auditChange{"permissionCodes": auditSetChange(before, after)})
	}
	return nil
}

func containsAllCoreAdminPermissions(codes []string) bool {
	set := make(map[string]bool, len(codes))
	for _, code := range codes {
		set[code] = true
	}
	for _, required := range []string{
		"system.users.view",
		"system.roles.permissions.update",
		"system.ldap.update",
		"system.audit.view",
		"system.audit.export",
	} {
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
