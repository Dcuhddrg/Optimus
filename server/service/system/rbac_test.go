package system

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRBACServiceAllowsRoleWithPermission(t *testing.T) {
	store := &fakeRBACStore{
		permissions: map[string]SystemPermissionSummary{
			"GET /api/v1/system/users": {ID: 10, Code: "system.users.view"},
		},
		allowed: map[uint]bool{1: true},
	}
	service := NewRBACService(store)

	allowed, err := service.AuthorizeAPI(context.Background(), "GET", "/api/v1/system/users", []uint{1})

	require.NoError(t, err)
	require.True(t, allowed)
}

func TestRBACServiceDeniesRoleWithoutPermission(t *testing.T) {
	store := &fakeRBACStore{
		permissions: map[string]SystemPermissionSummary{
			"GET /api/v1/system/users": {ID: 10, Code: "system.users.view"},
		},
		allowed: map[uint]bool{2: false},
	}
	service := NewRBACService(store)

	allowed, err := service.AuthorizeAPI(context.Background(), "GET", "/api/v1/system/users", []uint{2})

	require.NoError(t, err)
	require.False(t, allowed)
}

func TestRBACServiceDeniesUnseededPrivateAPI(t *testing.T) {
	service := NewRBACService(&fakeRBACStore{permissions: map[string]SystemPermissionSummary{}})

	allowed, err := service.AuthorizeAPI(context.Background(), "DELETE", "/api/v1/system/unknown", []uint{1})

	require.NoError(t, err)
	require.False(t, allowed)
}

type fakeRBACStore struct {
	permissions map[string]SystemPermissionSummary
	allowed     map[uint]bool
}

func (s *fakeRBACStore) FindAPIPermission(_ context.Context, method string, path string) (SystemPermissionSummary, bool, error) {
	permission, ok := s.permissions[method+" "+path]
	return permission, ok, nil
}

func (s *fakeRBACStore) RoleHasPermission(_ context.Context, roleID uint, _ uint) (bool, error) {
	return s.allowed[roleID], nil
}
