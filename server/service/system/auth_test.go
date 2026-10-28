package system_test

import (
	"context"
	"testing"
	"time"

	"optimus/server/config"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceLoginAdminSuccess(t *testing.T) {
	passwordHash, err := hashTestPassword("admin-secret")
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
	service := systemService.NewAuthService(store, fakeTokenIssuer{}, nil)

	result, err := service.Login(context.Background(), systemModel.LoginRequest{
		Username: "admin",
		Password: "admin-secret",
	}, systemService.AuditMeta{IP: "127.0.0.1", UserAgent: "test"})

	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Equal(t, "admin", result.User.Username)
}

// --- fakes ---

type fakeAuthStore struct {
	user             systemModel.SystemUser
	roles            []systemModel.SystemRoleSummary
	menuPermissions  []string
	audits           []systemService.AuditEvent
	lastLoginUpdated bool
	ldapConfig       systemModel.SystemLDAPConfig
}

func newFakeAuthStore(user systemModel.SystemUser) *fakeAuthStore {
	return &fakeAuthStore{
		user: user,
		roles: []systemModel.SystemRoleSummary{
			{ID: 1, Code: "admin", Name: "Administrator"},
		},
		menuPermissions: []string{"dashboard.view", "system.users.menu"},
	}
}

func (s *fakeAuthStore) FindUserByUsername(_ context.Context, username string) (systemModel.SystemUser, error) {
	if username != s.user.Username {
		return systemModel.SystemUser{}, systemService.ErrInvalidCredentials
	}
	return s.user, nil
}

func (s *fakeAuthStore) FindUserByID(_ context.Context, id uint) (systemModel.SystemUser, error) {
	if id != s.user.ID {
		return systemModel.SystemUser{}, systemService.ErrUserNotFound
	}
	return s.user, nil
}

func (s *fakeAuthStore) RolesForUser(_ context.Context, _ uint) ([]systemModel.SystemRoleSummary, error) {
	return s.roles, nil
}

func (s *fakeAuthStore) MenuPermissionCodesForRoles(_ context.Context, _ []uint) ([]string, error) {
	return s.menuPermissions, nil
}

func (s *fakeAuthStore) UpdateLastLogin(_ context.Context, _ uint, _ time.Time) error {
	s.lastLoginUpdated = true
	return nil
}

func (s *fakeAuthStore) WriteAudit(_ context.Context, event systemService.AuditEvent) error {
	s.audits = append(s.audits, event)
	return nil
}

func (s *fakeAuthStore) GetLDAPConfig(_ context.Context) (systemModel.SystemLDAPConfig, error) {
	return s.ldapConfig, nil
}

func (s *fakeAuthStore) SyncLDAPUser(_ context.Context, user systemService.LDAPUser) (systemModel.SystemUser, error) {
	return systemModel.SystemUser{}, nil
}

type fakeTokenIssuer struct{}

func (f fakeTokenIssuer) IssueToken(subject systemService.TokenSubject) (string, error) {
	return "fake-token-" + subject.Username, nil
}

func hashTestPassword(password string) (string, error) {
	return systemService.HashPassword(password)
}

func testAuthJWTConfig() config.JWT {
	return config.JWT{
		SigningKey:  "auth-test-signing-key",
		ExpiresTime: 3600,
		BufferTime:  300,
		Issuer:      "optimus-test",
	}
}
