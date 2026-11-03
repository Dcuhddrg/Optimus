package system

import (
	"context"

	"optimus/server/global"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

// Authorizer is service-owned to avoid importing HTTP middleware into Service.
type Authorizer interface {
	AuthorizeAPI(context.Context, string, string, []uint) (bool, error)
}

type RBACService struct{}

type SystemPermissionSummary struct {
	ID   uint
	Code string
}

func (s *RBACService) AuthorizeAPI(ctx context.Context, method string, path string, roleIDs []uint) (bool, error) {
	permission, found, err := s.findAPIPermission(ctx, method, path)
	if err != nil || !found {
		return false, err
	}

	for _, roleID := range roleIDs {
		allowed, err := s.roleHasPermission(ctx, roleID, permission.ID)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}

	return false, nil
}

func (s *RBACService) findAPIPermission(ctx context.Context, method string, path string) (SystemPermissionSummary, bool, error) {
	var permission systemModel.SystemPermission
	err := global.OPTIMUS_DB.WithContext(ctx).
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

func (s *RBACService) roleHasPermission(ctx context.Context, roleID uint, permissionID uint) (bool, error) {
	var count int64
	err := global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemRolePermission{}).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Count(&count).Error
	return count > 0, err
}
