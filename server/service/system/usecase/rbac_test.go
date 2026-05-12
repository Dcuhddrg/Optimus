package usecase_test

import (
	"context"
	"testing"

	"optimus/server/service/system/usecase"

	"github.com/stretchr/testify/require"
)

func TestRBACServiceAllowsRoleWithPermission(t *testing.T) {
	store := &fakeRBACStore{
		permissions: map[string]usecase.SystemPermissionSummary{
			"GET /api/v1/system/users": {ID: 10, Code: "system.users.view"},
		},
		allowed: map[uint]bool{1: true},
	}
	service := usecase.NewRBACService(store)

	allowed, err := service.AuthorizeAPI(context.Background(), "GET", "/api/v1/system/users", []uint{1})

	require.NoError(t, err)
	require.True(t, allowed)
}

type fakeRBACStore struct {
	permissions map[string]usecase.SystemPermissionSummary
	allowed     map[uint]bool
}

func (s *fakeRBACStore) FindAPIPermission(_ context.Context, method, path string) (usecase.SystemPermissionSummary, bool, error) {
	p, ok := s.permissions[method+" "+path]
	return p, ok, nil
}

func (s *fakeRBACStore) RoleHasPermission(_ context.Context, roleID, _ uint) (bool, error) {
	return s.allowed[roleID], nil
}
