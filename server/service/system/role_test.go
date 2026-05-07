package system

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestRoleServiceCreateRoleWritesAudit(t *testing.T) {
	store := &fakeRoleStore{}
	service := NewRoleService(store)

	role, err := service.CreateRole(context.Background(), 1, RoleRequest{
		Code:        "ops",
		Name:        "Operations",
		Description: "Ops role",
	}, AuditMeta{})

	require.NoError(t, err)
	require.Equal(t, "ops", role.Code)
	require.Equal(t, []string{"role.create"}, store.auditActions())
}

func TestRoleServiceRejectsDisablingAdminRole(t *testing.T) {
	store := &fakeRoleStore{roles: []systemModel.SystemRole{{ID: 1, Code: "admin", Status: "enabled"}}}
	service := NewRoleService(store)

	_, err := service.UpdateRoleStatus(context.Background(), 1, 1, "disabled", AuditMeta{})

	require.ErrorIs(t, err, ErrProtectedAdminRole)
	require.Empty(t, store.auditActions())
}

func TestRoleServiceSavePermissionsByCode(t *testing.T) {
	store := &fakeRoleStore{
		roles: []systemModel.SystemRole{{ID: 2, Code: "ops", Status: "enabled"}},
		permissionsByCode: map[string]systemModel.SystemPermission{
			"system.users.view": {ID: 10, Code: "system.users.view"},
		},
	}
	service := NewRoleService(store)

	err := service.UpdateRolePermissions(context.Background(), 1, 2, []string{"system.users.view"}, AuditMeta{})

	require.NoError(t, err)
	require.Equal(t, []uint{10}, store.savedPermissionIDs)
	require.Equal(t, []string{"role.permissions.update"}, store.auditActions())
}

func TestRoleServiceRejectsRemovingAdminCorePermissions(t *testing.T) {
	store := &fakeRoleStore{
		roles: []systemModel.SystemRole{{ID: 1, Code: "admin", Status: "enabled"}},
		permissionsByCode: map[string]systemModel.SystemPermission{
			"system.users.view": {ID: 10, Code: "system.users.view"},
		},
	}
	service := NewRoleService(store)

	err := service.UpdateRolePermissions(context.Background(), 1, 1, []string{"system.users.view"}, AuditMeta{})

	require.ErrorIs(t, err, ErrProtectedAdminRole)
	require.Empty(t, store.auditActions())
}

type fakeRoleStore struct {
	roles              []systemModel.SystemRole
	permissionsByCode  map[string]systemModel.SystemPermission
	audits             []AuditEvent
	savedPermissionIDs []uint
	nextRoleID         uint
}

func (s *fakeRoleStore) ListRoles(ctx context.Context) ([]systemModel.SystemRole, error) {
	return s.roles, nil
}

func (s *fakeRoleStore) FindRoleByID(ctx context.Context, id uint) (systemModel.SystemRole, error) {
	for _, role := range s.roles {
		if role.ID == id {
			return role, nil
		}
	}
	return systemModel.SystemRole{}, ErrRoleNotFound
}

func (s *fakeRoleStore) CreateRole(ctx context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error) {
	if s.nextRoleID == 0 {
		s.nextRoleID = 100
	}
	role.ID = s.nextRoleID
	s.roles = append(s.roles, role)
	return role, nil
}

func (s *fakeRoleStore) UpdateRole(ctx context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error) {
	for i := range s.roles {
		if s.roles[i].ID == role.ID {
			s.roles[i] = role
			return role, nil
		}
	}
	return role, nil
}

func (s *fakeRoleStore) PermissionsByCodes(ctx context.Context, codes []string) ([]systemModel.SystemPermission, error) {
	permissions := make([]systemModel.SystemPermission, 0, len(codes))
	for _, code := range codes {
		if permission, ok := s.permissionsByCode[code]; ok {
			permissions = append(permissions, permission)
		}
	}
	return permissions, nil
}

func (s *fakeRoleStore) ReplaceRolePermissions(ctx context.Context, roleID uint, permissionIDs []uint) error {
	s.savedPermissionIDs = append([]uint(nil), permissionIDs...)
	return nil
}

func (s *fakeRoleStore) WriteAudit(ctx context.Context, event AuditEvent) error {
	s.audits = append(s.audits, event)
	return nil
}

func (s *fakeRoleStore) auditActions() []string {
	actions := make([]string, 0, len(s.audits))
	for _, audit := range s.audits {
		actions = append(actions, audit.Action)
	}
	return actions
}
