package system_test

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
)

func TestUserServiceListUsersOmitsPasswordHash(t *testing.T) {
	store := &fakeUserStore{
		users: []systemModel.SystemUser{
			{ID: 1, Username: "admin", DisplayName: "Administrator", Source: "local", Status: "enabled", PasswordHash: "secret-hash"},
		},
	}
	service := systemService.NewUserService(store)

	result, err := service.ListUsers(context.Background())

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.Equal(t, "admin", result.Items[0].Username)
}

type fakeUserStore struct {
	users                   []systemModel.SystemUser
	audits                  []systemService.AuditEvent
	adminCapableAfterStatus bool
	adminCapableAfterRoles  bool
}

func (s *fakeUserStore) ListUsers(_ context.Context) ([]systemModel.SystemUser, error) {
	return s.users, nil
}

func (s *fakeUserStore) FindUserByID(_ context.Context, id uint) (systemModel.SystemUser, error) {
	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return systemModel.SystemUser{}, systemService.ErrUserNotFound
}

func (s *fakeUserStore) RolesForUser(_ context.Context, _ uint) ([]systemModel.SystemRoleSummary, error) {
	return nil, nil
}

func (s *fakeUserStore) UpdateUser(_ context.Context, user systemModel.SystemUser) (systemModel.SystemUser, error) {
	return user, nil
}

func (s *fakeUserStore) WouldHaveAdminCapableUserAfterStatus(_ context.Context, _ uint, _ string) (bool, error) {
	return s.adminCapableAfterStatus, nil
}

func (s *fakeUserStore) WouldHaveAdminCapableUserAfterRoles(_ context.Context, _ uint, _ []uint) (bool, error) {
	return s.adminCapableAfterRoles, nil
}

func (s *fakeUserStore) ReplaceUserRoles(_ context.Context, _ uint, _ []uint) error {
	return nil
}

func (s *fakeUserStore) WriteAudit(_ context.Context, event systemService.AuditEvent) error {
	s.audits = append(s.audits, event)
	return nil
}
