package system

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestUserServiceListUsersOmitsPasswordHash(t *testing.T) {
	store := &fakeUserStore{
		users: []systemModel.SystemUser{
			{ID: 1, Username: "admin", DisplayName: "Administrator", Source: "local", Status: "enabled", PasswordHash: "secret-hash"},
		},
		rolesByUser: map[uint][]SystemRoleSummary{
			1: {{ID: 1, Code: "admin", Name: "Administrator"}},
		},
	}
	service := NewUserService(store)

	result, err := service.ListUsers(context.Background())

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.Equal(t, "admin", result.Items[0].Username)
	require.Equal(t, []string{"admin"}, result.Items[0].Roles)
}

func TestUserServiceUpdateUserWritesAudit(t *testing.T) {
	store := &fakeUserStore{
		users: []systemModel.SystemUser{{ID: 2, Username: "ldap.user", Source: "ldap", Status: "enabled"}},
	}
	service := NewUserService(store)

	updated, err := service.UpdateUser(context.Background(), 1, 2, UpdateUserRequest{
		DisplayName: "LDAP User",
		Email:       "ldap@example.com",
	}, AuditMeta{})

	require.NoError(t, err)
	require.Equal(t, "LDAP User", updated.DisplayName)
	require.Equal(t, "ldap@example.com", updated.Email)
	require.Equal(t, []string{"user.update"}, store.auditActions())
}

func TestUserServiceRejectsDisablingLastAdminCapableUser(t *testing.T) {
	store := &fakeUserStore{
		users:                   []systemModel.SystemUser{{ID: 1, Username: "admin", Source: "local", Status: "enabled"}},
		adminCapableAfterStatus: false,
	}
	service := NewUserService(store)

	_, err := service.UpdateUserStatus(context.Background(), 1, 1, "disabled", AuditMeta{})

	require.ErrorIs(t, err, ErrLastAdminCapableUser)
	require.Empty(t, store.auditActions())
}

func TestUserServiceUpdateRolesRejectsRemovingLastAdminRole(t *testing.T) {
	store := &fakeUserStore{
		users:                  []systemModel.SystemUser{{ID: 1, Username: "admin", Source: "local", Status: "enabled"}},
		adminCapableAfterRoles: false,
	}
	service := NewUserService(store)

	err := service.UpdateUserRoles(context.Background(), 1, 1, []uint{2}, AuditMeta{})

	require.ErrorIs(t, err, ErrLastAdminCapableUser)
	require.Empty(t, store.auditActions())
}

type fakeUserStore struct {
	users                   []systemModel.SystemUser
	rolesByUser             map[uint][]SystemRoleSummary
	audits                  []AuditEvent
	adminCapableAfterStatus bool
	adminCapableAfterRoles  bool
	updatedUser             systemModel.SystemUser
}

func (s *fakeUserStore) ListUsers(ctx context.Context) ([]systemModel.SystemUser, error) {
	return s.users, nil
}

func (s *fakeUserStore) FindUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error) {
	for _, user := range s.users {
		if user.ID == id {
			return user, nil
		}
	}
	return systemModel.SystemUser{}, ErrUserNotFound
}

func (s *fakeUserStore) RolesForUser(ctx context.Context, userID uint) ([]SystemRoleSummary, error) {
	return s.rolesByUser[userID], nil
}

func (s *fakeUserStore) UpdateUser(ctx context.Context, user systemModel.SystemUser) (systemModel.SystemUser, error) {
	s.updatedUser = user
	for i := range s.users {
		if s.users[i].ID == user.ID {
			s.users[i] = user
			return user, nil
		}
	}
	return user, nil
}

func (s *fakeUserStore) WouldHaveAdminCapableUserAfterStatus(ctx context.Context, userID uint, status string) (bool, error) {
	return s.adminCapableAfterStatus, nil
}

func (s *fakeUserStore) WouldHaveAdminCapableUserAfterRoles(ctx context.Context, userID uint, roleIDs []uint) (bool, error) {
	return s.adminCapableAfterRoles, nil
}

func (s *fakeUserStore) ReplaceUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return nil
}

func (s *fakeUserStore) WriteAudit(ctx context.Context, event AuditEvent) error {
	s.audits = append(s.audits, event)
	return nil
}

func (s *fakeUserStore) auditActions() []string {
	actions := make([]string, 0, len(s.audits))
	for _, audit := range s.audits {
		actions = append(actions, audit.Action)
	}
	return actions
}
