package store

import (
	"context"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"gorm.io/gorm"
)

type GormRBACStore struct {
	db *gorm.DB
}

func NewGormRBACStore(db *gorm.DB) *GormRBACStore {
	return &GormRBACStore{db: db}
}

func (s *GormRBACStore) FindAPIPermission(ctx context.Context, method string, path string) (systemService.SystemPermissionSummary, bool, error) {
	var permission systemModel.SystemPermission
	err := s.db.WithContext(ctx).
		Where("type = ? AND method = ? AND path = ?", "api", method, path).
		First(&permission).Error
	if err == nil {
		return systemService.SystemPermissionSummary{ID: permission.ID, Code: permission.Code}, true, nil
	}
	if err == gorm.ErrRecordNotFound {
		return systemService.SystemPermissionSummary{}, false, nil
	}
	return systemService.SystemPermissionSummary{}, false, err
}

func (s *GormRBACStore) RoleHasPermission(ctx context.Context, roleID uint, permissionID uint) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&systemModel.SystemRolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}
