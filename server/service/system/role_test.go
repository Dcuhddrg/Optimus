package system_test

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
)

func TestRoleServiceCreateRoleWritesAudit(t *testing.T) {
	store := &fakeRoleStore{}
	service := systemService.NewRoleService(store)

	role, err := service.CreateRole(context.Background(), 1, systemModel.RoleRequest{
		Code: "ops",
		Name: "Operations",
	}, systemService.AuditMeta{})

	require.NoError(t, err)
	require.Equal(t, "ops", role.Code)
}

type fakeRoleStore struct {
	roles             []systemModel.SystemRole
	permissionsByCode map[string]systemModel.SystemPermission
	audits            []systemService.AuditEvent
	savedPermIDs      []uint
}

func (s *fakeRoleStore) ListRoles(_ context.Context) ([]systemModel.SystemRole, error) {
	return s.roles, nil
}

func (s *fakeRoleStore) FindRoleByID(_ context.Context, id uint) (systemModel.SystemRole, error) {
	for _, r := range s.roles {
		if r.ID == id {
			return r, nil
		}
	}
	return systemModel.SystemRole{}, systemService.ErrRoleNotFound
}

func (s *fakeRoleStore) CreateRole(_ context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error) {
	role.ID = 100
	s.roles = append(s.roles, role)
	return role, nil
}

func (s *fakeRoleStore) UpdateRole(_ context.Context, role systemModel.SystemRole) (systemModel.SystemRole, error) {
	return role, nil
}

func (s *fakeRoleStore) PermissionsByCodes(_ context.Context, codes []string) ([]systemModel.SystemPermission, error) {
	var perms []systemModel.SystemPermission
	for _, code := range codes {
		if p, ok := s.permissionsByCode[code]; ok {
			perms = append(perms, p)
		}
	}
	return perms, nil
}

func (s *fakeRoleStore) ReplaceRolePermissions(_ context.Context, _ uint, permIDs []uint) error {
	s.savedPermIDs = permIDs
	return nil
}

func (s *fakeRoleStore) WriteAudit(_ context.Context, event systemService.AuditEvent) error {
	s.audits = append(s.audits, event)
	return nil
}
