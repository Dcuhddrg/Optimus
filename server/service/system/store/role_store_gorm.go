package store

import (
	"context"
	"errors"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"gorm.io/gorm"
)

type GormRoleStore struct {
	db *gorm.DB
}

func NewGormRoleStore(db *gorm.DB) *GormRoleStore {
	return &GormRoleStore{db: db}
}

func (s *GormRoleStore) ListRoles(ctx context.Context) ([]systemModel.SystemRole, error) {
	var roles []systemModel.SystemRole
	err := s.db.WithContext(ctx).Order("id ASC").Find(&roles).Error
	return roles, err
}

func (s *GormRoleStore) FindRoleByID(ctx context.Context, id uint) (systemModel.SystemRole, error) {
	var role systemModel.SystemRole
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemRole{}, systemService.ErrRoleNotFound
	}
	return role, err
}

func (s *GormRoleStore) CreateRole(ctx context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error) {
	err := s.db.WithContext(ctx).Create(&role).Error
	return role, err
}

func (s *GormRoleStore) UpdateRole(ctx context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error) {
	err := s.db.WithContext(ctx).Save(&role).Error
	return role, err
}

func (s *GormRoleStore) PermissionsByCodes(ctx context.Context, codes []string) ([]systemModel.SystemPermission, error) {
	var permissions []systemModel.SystemPermission
	err := s.db.WithContext(ctx).Where("code IN ?", codes).Find(&permissions).Error
	return permissions, err
}

func (s *GormRoleStore) ReplaceRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&systemModel.SystemRolePermission{}).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			if err := tx.Create(&systemModel.SystemRolePermission{RoleID: roleID, PermissionID: permissionID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *GormRoleStore) WriteAudit(ctx context.Context, event systemService.AuditEvent) error {
	return NewGormAuthStore(s.db).WriteAudit(ctx, event)
}
