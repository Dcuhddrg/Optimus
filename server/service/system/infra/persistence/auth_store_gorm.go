package persistence

import (
	"context"
	"time"

	"optimus/server/service/system/usecase"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

const (
	enabledUserStatus = "enabled"
	enabledRoleStatus = "enabled"
)

type GormAuthStore struct {
	db *gorm.DB
}

func NewGormAuthStore(db *gorm.DB) *GormAuthStore {
	return &GormAuthStore{db: db}
}

func (s *GormAuthStore) FindUserByUsername(ctx context.Context, username string) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return user, err
}

func (s *GormAuthStore) FindUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}

func (s *GormAuthStore) GetLDAPConfig(ctx context.Context) (systemModel.SystemLDAPConfig, error) {
	var cfg systemModel.SystemLDAPConfig
	err := s.db.WithContext(ctx).Order("id ASC").First(&cfg).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return systemModel.SystemLDAPConfig{}, nil
	}
	return cfg, err
}

func (s *GormAuthStore) SyncLDAPUser(ctx context.Context, ldapUser usecase.LDAPUser) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := s.db.WithContext(ctx).Where("username = ?", ldapUser.Username).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return systemModel.SystemUser{}, err
	}
	if err == gorm.ErrRecordNotFound {
		user = systemModel.SystemUser{
			UUID:     "ldap:" + ldapUser.Username,
			Username: ldapUser.Username,
			Source:   "ldap",
			Status:   enabledUserStatus,
		}
	}
	user.DisplayName = ldapUser.DisplayName
	user.Email = ldapUser.Email
	if err == gorm.ErrRecordNotFound {
		if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
			return systemModel.SystemUser{}, err
		}
		if err := s.bindDefaultUserRole(ctx, user.ID); err != nil {
			return systemModel.SystemUser{}, err
		}
		return user, nil
	}
	return user, s.db.WithContext(ctx).Save(&user).Error
}

func (s *GormAuthStore) bindDefaultUserRole(ctx context.Context, userID uint) error {
	var role systemModel.SystemRole
	if err := s.db.WithContext(ctx).Where("code = ?", "user").First(&role).Error; err != nil {
		return err
	}
	userRole := systemModel.SystemUserRole{UserID: userID, RoleID: role.ID}
	return s.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, role.ID).FirstOrCreate(&userRole).Error
}

func (s *GormAuthStore) RolesForUser(ctx context.Context, userID uint) ([]usecase.SystemRoleSummary, error) {
	var roles []usecase.SystemRoleSummary
	err := s.db.WithContext(ctx).
		Table((&systemModel.SystemRole{}).TableName()+" AS roles").
		Select("roles.id, roles.code, roles.name").
		Joins("JOIN sys_user_roles user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, enabledRoleStatus).
		Order("roles.id ASC").
		Scan(&roles).Error
	return roles, err
}

func (s *GormAuthStore) MenuPermissionCodesForRoles(ctx context.Context, roleIDs []uint) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	var codes []string
	err := s.db.WithContext(ctx).
		Table((&systemModel.SystemPermission{}).TableName()+" AS permissions").
		Select("permissions.code").
		Joins("JOIN sys_role_permissions role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id IN ? AND permissions.type = ?", roleIDs, "menu").
		Group("permissions.code").
		Order("MIN(permissions.sort) ASC").
		Pluck("permissions.code", &codes).Error
	return codes, err
}

func (s *GormAuthStore) UpdateLastLogin(ctx context.Context, userID uint, at time.Time) error {
	return s.db.WithContext(ctx).
		Model(&systemModel.SystemUser{}).
		Where("id = ?", userID).
		Update("last_login_at", at).Error
}

func (s *GormAuthStore) WriteAudit(ctx context.Context, event usecase.AuditEvent) error {
	log := systemModel.SystemAuditLog{
		ActorUserID:   event.ActorUserID,
		ActorUsername: event.ActorUsername,
		Action:        event.Action,
		ResourceType:  event.ResourceType,
		ResourceID:    event.ResourceID,
		Result:        event.Result,
		IP:            event.IP,
		UserAgent:     event.UserAgent,
		Detail:        event.Detail,
	}
	return s.db.WithContext(ctx).Create(&log).Error
}
