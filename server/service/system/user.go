package system

import (
	"context"
	"errors"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	// ErrBuiltinAdminProtected: the built-in admin account must stay enabled
	// and keep the admin role, so the system always has an administrator.
	ErrBuiltinAdminProtected = errors.New("built-in admin must stay enabled with the admin role")
)

type UserService struct{}

func (s *UserService) ListUsers(ctx context.Context) (systemResponse.UserListResult, error) {
	users, err := s.listUsers(ctx)
	if err != nil {
		return systemResponse.UserListResult{}, err
	}

	items := make([]systemResponse.UserListItem, 0, len(users))
	for _, user := range users {
		roles, err := queryUserRoles(ctx, user.ID)
		if err != nil {
			return systemResponse.UserListResult{}, err
		}
		_, roleCodes := summarizeRoles(roles)
		items = append(items, systemResponse.UserListItem{
			SystemUserSummary: summarizeUser(user),
			Roles:             roleCodes,
		})
	}

	return systemResponse.UserListResult{Items: items, Total: len(items)}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, actorID uint, userID uint, req systemRequest.UpdateUserRequest, meta AuditMeta) (systemResponse.SystemUserSummary, error) {
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		return systemResponse.SystemUserSummary{}, err
	}
	// Write only the edited columns: a full-row Save of this earlier read
	// would restore a status disabled in the meantime.
	updated, err := s.updateUserColumns(ctx, user.ID, map[string]any{"display_name": req.DisplayName, "email": req.Email})
	if err != nil {
		return systemResponse.SystemUserSummary{}, err
	}
	recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "user.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, actorID uint, userID uint, status string, meta AuditMeta) (systemResponse.SystemUserSummary, error) {
	event := failedAudit(actorID, meta, "user.status.update", "user", userID)
	defer func() { recordAudit(ctx, event) }()
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return systemResponse.SystemUserSummary{}, err
	}
	if isBuiltinAdmin(user) && status != DefaultUserStatus {
		event.Detail = auditDetails("builtin_admin_protected", nil)
		return systemResponse.SystemUserSummary{}, ErrBuiltinAdminProtected
	}
	before := user.Status
	if err := s.writeUserColumns(ctx, user.ID, map[string]any{"status": status}); err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return systemResponse.SystemUserSummary{}, err
	}
	// Committed: invalidate before anything else can fail, so a disable
	// takes effect even if re-reading the row below fails.
	InvalidateUserAuthState(userID)
	updated, err := s.findUserByID(ctx, userID)
	if err != nil {
		event.Result = "success"
		event.Detail = auditDetails("snapshot_unavailable", nil)
		return systemResponse.SystemUserSummary{}, err
	}
	event.Result = "success"
	event.Detail = auditDetails("updated", map[string]auditChange{"status": {Before: before, After: updated.Status}})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserRoles(ctx context.Context, actorID uint, userID uint, roleIDs []uint, meta AuditMeta) error {
	event := failedAudit(actorID, meta, "user.roles.update", "user", userID)
	defer func() { recordAudit(ctx, event) }()
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return err
	}
	if isBuiltinAdmin(user) {
		keepsAdmin, err := includesAdminRole(ctx, roleIDs)
		if err != nil {
			event.Detail = auditDetails("admin_check_failed", nil)
			return err
		}
		if !keepsAdmin {
			event.Detail = auditDetails("builtin_admin_protected", nil)
			return ErrBuiltinAdminProtected
		}
	}
	before, beforeErr := queryUserRoles(ctx, userID)
	if err := s.replaceUserRoles(ctx, userID, roleIDs); err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return err
	}
	InvalidateUserAuthState(userID)
	after, afterErr := queryUserRoles(ctx, userID)
	event.Result = "success"
	if beforeErr != nil || afterErr != nil {
		event.Detail = auditDetails("snapshot_unavailable", nil)
	} else {
		beforeIDs, afterIDs := make([]uint, 0, len(before)), make([]uint, 0, len(after))
		for _, role := range before {
			beforeIDs = append(beforeIDs, role.ID)
		}
		for _, role := range after {
			afterIDs = append(afterIDs, role.ID)
		}
		event.Detail = auditDetails("updated", map[string]auditChange{"roleIds": auditSetChange(beforeIDs, afterIDs)})
	}
	return nil
}

func isBuiltinAdmin(user systemModel.SystemUser) bool {
	return user.Source == "local" && user.UUID == BuiltinAdminUUID
}

func includesAdminRole(ctx context.Context, roleIDs []uint) (bool, error) {
	var adminRole systemModel.SystemRole
	if err := global.OPTIMUS_DB.WithContext(ctx).Where("code = ?", AdminRoleCode).First(&adminRole).Error; err != nil {
		return false, err
	}
	for _, roleID := range roleIDs {
		if roleID == adminRole.ID {
			return true, nil
		}
	}
	return false, nil
}

func (s *UserService) listUsers(ctx context.Context) ([]systemModel.SystemUser, error) {
	var users []systemModel.SystemUser
	err := global.OPTIMUS_DB.WithContext(ctx).Order("id ASC").Find(&users).Error
	return users, err
}

func (s *UserService) findUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := global.OPTIMUS_DB.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemUser{}, ErrUserNotFound
	}
	return user, err
}

// updateUserColumns updates only the given columns and returns the row as
// stored afterwards. Never Save a whole user read earlier: concurrent writes
// to other columns (status, roles' holders) would be overwritten.
func (s *UserService) updateUserColumns(ctx context.Context, userID uint, columns map[string]any) (systemModel.SystemUser, error) {
	if err := s.writeUserColumns(ctx, userID, columns); err != nil {
		return systemModel.SystemUser{}, err
	}
	return s.findUserByID(ctx, userID)
}

func (s *UserService) writeUserColumns(ctx context.Context, userID uint, columns map[string]any) error {
	return global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemUser{}).Where("id = ?", userID).Updates(columns).Error
}

func (s *UserService) replaceUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return global.OPTIMUS_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&systemModel.SystemUserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&systemModel.SystemUserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
