package persistence

import (
	"context"
	"errors"

	"optimus/server/service/system/usecase"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

type GormUserStore struct {
	db *gorm.DB
}

func NewGormUserStore(db *gorm.DB) *GormUserStore {
	return &GormUserStore{db: db}
}

func (s *GormUserStore) ListUsers(ctx context.Context) ([]systemModel.SystemUser, error) {
	var users []systemModel.SystemUser
	err := s.db.WithContext(ctx).Order("id ASC").Find(&users).Error
	return users, err
}

func (s *GormUserStore) FindUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemUser{}, usecase.ErrUserNotFound
	}
	return user, err
}

func (s *GormUserStore) RolesForUser(ctx context.Context, userID uint) ([]usecase.SystemRoleSummary, error) {
	return NewGormAuthStore(s.db).RolesForUser(ctx, userID)
}

func (s *GormUserStore) UpdateUser(ctx context.Context, user systemModel.SystemUser) (systemModel.SystemUser, error) {
	err := s.db.WithContext(ctx).Save(&user).Error
	return user, err
}

func (s *GormUserStore) WouldHaveAdminCapableUserAfterStatus(ctx context.Context, userID uint, status string) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Table("sys_users AS users").
		Joins("JOIN sys_user_roles user_roles ON user_roles.user_id = users.id").
		Joins("JOIN sys_roles roles ON roles.id = user_roles.role_id").
		Where("roles.code = ? AND roles.status = ?", "admin", enabledRoleStatus).
		Where("(users.id <> ? AND users.status = ?) OR (users.id = ? AND ? = ?)", userID, enabledUserStatus, userID, status, enabledUserStatus).
		Count(&count).Error
	return count > 0, err
}

func (s *GormUserStore) WouldHaveAdminCapableUserAfterRoles(ctx context.Context, userID uint, roleIDs []uint) (bool, error) {
	var adminRole systemModel.SystemRole
	if err := s.db.WithContext(ctx).Where("code = ? AND status = ?", "admin", enabledRoleStatus).First(&adminRole).Error; err != nil {
		return false, err
	}

	hasAdminRole := false
	for _, roleID := range roleIDs {
		if roleID == adminRole.ID {
			hasAdminRole = true
			break
		}
	}

	var count int64
	err := s.db.WithContext(ctx).
		Table("sys_users AS users").
		Joins("JOIN sys_user_roles user_roles ON user_roles.user_id = users.id").
		Joins("JOIN sys_roles roles ON roles.id = user_roles.role_id").
		Where("users.id <> ? AND users.status = ? AND roles.code = ? AND roles.status = ?", userID, enabledUserStatus, "admin", enabledRoleStatus).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	var user systemModel.SystemUser
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}
	return user.Status == enabledUserStatus && hasAdminRole, nil
}

func (s *GormUserStore) ReplaceUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (s *GormUserStore) WriteAudit(ctx context.Context, event usecase.AuditEvent) error {
	return NewGormAuthStore(s.db).WriteAudit(ctx, event)
}
