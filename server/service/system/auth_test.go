package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"optimus/server/config"
	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceLoginAdminSuccess(t *testing.T) {
	passwordHash, err := HashPassword("admin-secret")
	require.NoError(t, err)
	store := newFakeAuthStore(systemModel.SystemUser{
		ID:           1,
		UUID:         "admin-uuid",
		Username:     "admin",
		DisplayName:  "Administrator",
		Source:       "local",
		Status:       "enabled",
		PasswordHash: passwordHash,
	})
	service := NewAuthService(store, testAuthJWTConfig())

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "admin",
		Password: "admin-secret",
	}, AuditMeta{IP: "127.0.0.1", UserAgent: "test"})

	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Equal(t, "admin", result.User.Username)
	require.Equal(t, []string{"admin"}, result.Roles)
	require.Equal(t, []string{"dashboard.view", "system.users.menu"}, result.MenuPermissions)
	require.Equal(t, []string{"auth.login.success"}, store.auditActions())
	require.True(t, store.lastLoginUpdated)
}

func TestAuthServiceLoginRejectsWrongPassword(t *testing.T) {
	passwordHash, err := HashPassword("admin-secret")
	require.NoError(t, err)
	store := newFakeAuthStore(systemModel.SystemUser{
		ID:           1,
		UUID:         "admin-uuid",
		Username:     "admin",
		Source:       "local",
		Status:       "enabled",
		PasswordHash: passwordHash,
	})
	service := NewAuthService(store, testAuthJWTConfig())

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "admin",
		Password: "wrong-secret",
	}, AuditMeta{})

	require.ErrorIs(t, err, ErrInvalidCredentials)
	require.Empty(t, result.Token)
	require.Equal(t, []string{"auth.login.failure"}, store.auditActions())
}

func TestAuthServiceLoginRejectsDisabledUser(t *testing.T) {
	passwordHash, err := HashPassword("admin-secret")
	require.NoError(t, err)
	store := newFakeAuthStore(systemModel.SystemUser{
		ID:           1,
		UUID:         "admin-uuid",
		Username:     "admin",
		Source:       "local",
		Status:       "disabled",
		PasswordHash: passwordHash,
	})
	service := NewAuthService(store, testAuthJWTConfig())

	result, err := service.Login(context.Background(), LoginRequest{
		Username: "admin",
		Password: "admin-secret",
	}, AuditMeta{})

	require.ErrorIs(t, err, ErrUserDisabled)
	require.Empty(t, result.Token)
	require.Equal(t, []string{"auth.login.failure"}, store.auditActions())
}

func TestAuthServiceCurrentUser(t *testing.T) {
	store := newFakeAuthStore(systemModel.SystemUser{
		ID:          1,
		UUID:        "admin-uuid",
		Username:    "admin",
		DisplayName: "Administrator",
		Source:      "local",
		Status:      "enabled",
	})
	service := NewAuthService(store, testAuthJWTConfig())

	result, err := service.CurrentUser(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, "admin", result.User.Username)
	require.Equal(t, []string{"admin"}, result.Roles)
	require.Equal(t, []string{"dashboard.view", "system.users.menu"}, result.MenuPermissions)
}

func TestAuthServiceLogoutWritesAudit(t *testing.T) {
	store := newFakeAuthStore(systemModel.SystemUser{ID: 1, Username: "admin"})
	service := NewAuthService(store, testAuthJWTConfig())

	err := service.Logout(context.Background(), 1, AuditMeta{IP: "127.0.0.1", UserAgent: "test"})

	require.NoError(t, err)
	require.Equal(t, []string{"auth.logout"}, store.auditActions())
}

type fakeAuthStore struct {
	user             systemModel.SystemUser
	roles            []SystemRoleSummary
	menuPermissions  []string
	audits           []AuditEvent
	lastLoginUpdated bool
}

func newFakeAuthStore(user systemModel.SystemUser) *fakeAuthStore {
	return &fakeAuthStore{
		user: user,
		roles: []SystemRoleSummary{
			{ID: 1, Code: "admin", Name: "Administrator"},
		},
		menuPermissions: []string{"dashboard.view", "system.users.menu"},
	}
}

func (s *fakeAuthStore) FindUserByUsername(_ context.Context, username string) (systemModel.SystemUser, error) {
	if username != s.user.Username {
		return systemModel.SystemUser{}, ErrInvalidCredentials
	}
	return s.user, nil
}

func (s *fakeAuthStore) FindUserByID(_ context.Context, id uint) (systemModel.SystemUser, error) {
	if id != s.user.ID {
		return systemModel.SystemUser{}, errors.New("not found")
	}
	return s.user, nil
}

func (s *fakeAuthStore) RolesForUser(_ context.Context, _ uint) ([]SystemRoleSummary, error) {
	return s.roles, nil
}

func (s *fakeAuthStore) MenuPermissionCodesForRoles(_ context.Context, _ []uint) ([]string, error) {
	return s.menuPermissions, nil
}

func (s *fakeAuthStore) UpdateLastLogin(_ context.Context, _ uint, _ time.Time) error {
	s.lastLoginUpdated = true
	return nil
}

func (s *fakeAuthStore) WriteAudit(_ context.Context, event AuditEvent) error {
	s.audits = append(s.audits, event)
	return nil
}

func (s *fakeAuthStore) auditActions() []string {
	actions := make([]string, 0, len(s.audits))
	for _, audit := range s.audits {
		actions = append(actions, audit.Action)
	}
	return actions
}

func testAuthJWTConfig() config.JWT {
	return config.JWT{
		SigningKey:  "auth-test-signing-key",
		ExpiresTime: 3600,
		BufferTime:  300,
		Issuer:      "optimus-test",
	}
}
