package usecase

import "context"

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
