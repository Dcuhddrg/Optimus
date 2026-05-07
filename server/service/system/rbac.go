package system

import (
	"context"

	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

type RBACStore interface {
	FindAPIPermission(ctx context.Context, method string, path string) (SystemPermissionSummary, bool, error)
	RoleHasPermission(ctx context.Context, roleID uint, permissionID uint) (bool, error)
}

type RBACService struct {
	store RBACStore
}

type SystemPermissionSummary struct {
	ID   uint
	Code string
}

func NewRBACService(store RBACStore) *RBACService {
	return &RBACService{store: store}
}

func (s *RBACService) AuthorizeAPI(ctx context.Context, method string, path string, roleIDs []uint) (bool, error) {
	permission, found, err := s.store.FindAPIPermission(ctx, method, path)
	if err != nil || !found {
		return false, err
	}

	for _, roleID := range roleIDs {
		allowed, err := s.store.RoleHasPermission(ctx, roleID, permission.ID)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

type GormRBACStore struct {
	db *gorm.DB
}

func NewGormRBACStore(db *gorm.DB) *GormRBACStore {
	return &GormRBACStore{db: db}
}

func (s *GormRBACStore) FindAPIPermission(ctx context.Context, method string, path string) (SystemPermissionSummary, bool, error) {
	var permission systemModel.SystemPermission
	err := s.db.WithContext(ctx).
		Where("type = ? AND method = ? AND path = ?", "api", method, path).
		First(&permission).Error
	if err == nil {
		return SystemPermissionSummary{ID: permission.ID, Code: permission.Code}, true, nil
	}
	if err == gorm.ErrRecordNotFound {
		return SystemPermissionSummary{}, false, nil
	}
	return SystemPermissionSummary{}, false, err
}

func (s *GormRBACStore) RoleHasPermission(ctx context.Context, roleID uint, permissionID uint) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&systemModel.SystemRolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}
