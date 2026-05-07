package system

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestLDAPServiceConfigMasksBindPassword(t *testing.T) {
	store := &fakeLDAPStore{config: systemModel.SystemLDAPConfig{
		Enabled:         true,
		URL:             "ldap://example.test:389",
		BaseDN:          "dc=example,dc=test",
		BindDN:          "cn=admin,dc=example,dc=test",
		BindPassword:    "secret",
		UserFilter:      "(uid={username})",
		UsernameAttr:    "uid",
		DisplayNameAttr: "cn",
		EmailAttr:       "mail",
	}}
	service := NewLDAPService(store, &fakeLDAPClient{})

	config, err := service.GetConfig(context.Background())

	require.NoError(t, err)
	require.True(t, config.HasBindPassword)
	require.Empty(t, config.BindPassword)
}

func TestLDAPServiceSaveConfigPreservesEmptyBindPassword(t *testing.T) {
	store := &fakeLDAPStore{config: systemModel.SystemLDAPConfig{BindPassword: "saved-secret"}}
	service := NewLDAPService(store, &fakeLDAPClient{})

	_, err := service.SaveConfig(context.Background(), 1, LDAPConfigRequest{
		Enabled:         true,
		URL:             "ldap://example.test:389",
		BaseDN:          "dc=example,dc=test",
		BindDN:          "cn=admin,dc=example,dc=test",
		BindPassword:    "",
		UserFilter:      "(uid={username})",
		UsernameAttr:    "uid",
		DisplayNameAttr: "cn",
		EmailAttr:       "mail",
	}, AuditMeta{})

	require.NoError(t, err)
	require.Equal(t, "saved-secret", store.config.BindPassword)
	require.Equal(t, []string{"ldap.config.update"}, store.auditActions())
}

func TestLDAPServiceTestConfigUsesSavedBindPassword(t *testing.T) {
	store := &fakeLDAPStore{config: systemModel.SystemLDAPConfig{BindPassword: "saved-secret"}}
	client := fakeLDAPClient{}
	service := NewLDAPService(store, &client)

	result, err := service.TestConfig(context.Background(), 1, LDAPTestRequest{
		URL:                  "ldap://example.test:389",
		BaseDN:               "dc=example,dc=test",
		BindDN:               "cn=admin,dc=example,dc=test",
		UseSavedBindPassword: true,
		UserFilter:           "(uid={username})",
		UsernameAttr:         "uid",
		DisplayNameAttr:      "cn",
		EmailAttr:            "mail",
	}, AuditMeta{})

	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, "saved-secret", client.tested.BindPassword)
	require.Equal(t, []string{"ldap.test.success"}, store.auditActions())
}

func TestLDAPServiceTestConfigRequiresUsernamePlaceholder(t *testing.T) {
	store := &fakeLDAPStore{}
	service := NewLDAPService(store, &fakeLDAPClient{})

	_, err := service.TestConfig(context.Background(), 1, LDAPTestRequest{
		URL:        "ldap://example.test:389",
		BaseDN:     "dc=example,dc=test",
		BindDN:     "cn=admin,dc=example,dc=test",
		UserFilter: "(uid=static)",
	}, AuditMeta{})

	require.ErrorIs(t, err, ErrInvalidLDAPConfig)
	require.Equal(t, []string{"ldap.test.failure"}, store.auditActions())
}

type fakeLDAPStore struct {
	config systemModel.SystemLDAPConfig
	audits []AuditEvent
}

func (s *fakeLDAPStore) GetLDAPConfig(ctx context.Context) (systemModel.SystemLDAPConfig, error) {
	return s.config, nil
}

func (s *fakeLDAPStore) SaveLDAPConfig(ctx context.Context, cfg systemModel.SystemLDAPConfig) (systemModel.SystemLDAPConfig, error) {
	s.config = cfg
	return cfg, nil
}

func (s *fakeLDAPStore) WriteAudit(ctx context.Context, event AuditEvent) error {
	s.audits = append(s.audits, event)
	return nil
}

func (s *fakeLDAPStore) auditActions() []string {
	actions := make([]string, 0, len(s.audits))
	for _, audit := range s.audits {
		actions = append(actions, audit.Action)
	}
	return actions
}

type fakeLDAPClient struct {
	tested LDAPConnectionConfig
	err    error
}

func (c *fakeLDAPClient) Test(ctx context.Context, cfg LDAPConnectionConfig) error {
	c.tested = cfg
	return c.err
}
