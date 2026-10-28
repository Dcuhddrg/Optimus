package system_test

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

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
	service := systemService.NewLDAPService(store, &fakeLDAPClientUC{})

	config, err := service.GetConfig(context.Background())

	require.NoError(t, err)
	require.True(t, config.HasBindPassword)
	require.Empty(t, config.BindPassword)
}

type fakeLDAPStore struct {
	config systemModel.SystemLDAPConfig
	audits []systemService.AuditEvent
}

func (s *fakeLDAPStore) GetLDAPConfig(_ context.Context) (systemModel.SystemLDAPConfig, error) {
	return s.config, nil
}

func (s *fakeLDAPStore) SaveLDAPConfig(_ context.Context, cfg systemModel.SystemLDAPConfig) (systemModel.SystemLDAPConfig, error) {
	s.config = cfg
	return cfg, nil
}

func (s *fakeLDAPStore) WriteAudit(_ context.Context, event systemService.AuditEvent) error {
	s.audits = append(s.audits, event)
	return nil
}

type fakeLDAPClientUC struct{}

func (c *fakeLDAPClientUC) Test(_ context.Context, cfg systemService.LDAPConnectionConfig) error {
	return nil
}
