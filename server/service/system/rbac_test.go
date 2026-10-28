package system_test

import (
	"context"
	"testing"

	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
)

func TestRBACServiceAllowsRoleWithPermission(t *testing.T) {
	store := &fakeRBACStore{
		permissions: map[string]systemService.SystemPermissionSummary{
			"GET /api/v1/system/users": {ID: 10, Code: "system.users.view"},
		},
		allowed: map[uint]bool{1: true},
	}
	service := systemService.NewRBACService(store)

	allowed, err := service.AuthorizeAPI(context.Background(), "GET", "/api/v1/system/users", []uint{1})

	require.NoError(t, err)
	require.True(t, allowed)
}

type fakeRBACStore struct {
	permissions map[string]systemService.SystemPermissionSummary
	allowed     map[uint]bool
}

func (s *fakeRBACStore) FindAPIPermission(_ context.Context, method, path string) (systemService.SystemPermissionSummary, bool, error) {
	p, ok := s.permissions[method+" "+path]
	return p, ok, nil
}

func (s *fakeRBACStore) RoleHasPermission(_ context.Context, roleID, _ uint) (bool, error) {
	return s.allowed[roleID], nil
}
