package system_test

import (
	"context"
	"errors"
	"testing"

	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLDAPServiceConfigMasksBindPassword(t *testing.T) {
	db := openExternalPersistenceDB(t)
	cfg := systemModel.SystemLDAPConfig{
		Enabled:         true,
		URL:             "ldap://example.test:389",
		BaseDN:          "dc=example,dc=test",
		BindDN:          "cn=admin,dc=example,dc=test",
		BindPassword:    "secret",
		UserFilter:      "(uid={username})",
		UsernameAttr:    "uid",
		DisplayNameAttr: "cn",
		EmailAttr:       "mail",
	}
	require.NoError(t, db.Create(&cfg).Error)
	service := systemService.NewLDAPService(&fakeLDAPClientUC{})

	config, err := service.GetConfig(context.Background())

	require.NoError(t, err)
	require.True(t, config.HasBindPassword)
	require.Empty(t, config.BindPassword)
}

type failingLDAPClient struct{}

func (failingLDAPClient) Test(context.Context, systemService.LDAPConnectionConfig) error {
	return errors.New("DO_NOT_LOG_THIS")
}

func TestLDAPSnapshotFailureStillSavesConfig(t *testing.T) {
	db := openExternalPersistenceDB(t)
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("test:fail_ldap_read", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_ldap_config" {
			tx.AddError(errors.New("snapshot read failed"))
		}
	}))
	svc := systemService.NewLDAPService(&fakeLDAPClientUC{})
	_, err := svc.SaveConfig(context.Background(), 1, systemRequest.LDAPConfigRequest{Enabled: true, BindPassword: "SECRET"}, systemService.AuditMeta{})
	require.NoError(t, err)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "success", audit.Result)
	require.Contains(t, audit.Detail, "snapshot_unavailable")
	require.NotContains(t, audit.Detail, "SECRET")
}

func TestLDAPConnectionFailureAuditOmitsSecrets(t *testing.T) {
	db := openExternalPersistenceDB(t)
	svc := systemService.NewLDAPService(failingLDAPClient{})
	_, err := svc.TestConfig(context.Background(), 1, systemRequest.LDAPTestRequest{
		URL: "ldap://test", BindPassword: "PASSWORD_SECRET", UserFilter: "(uid={username})",
	}, systemService.AuditMeta{})
	require.Error(t, err)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "ldap.test.failure").First(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "connection_test_failed")
	require.NotContains(t, audit.Detail, "DO_NOT_LOG_THIS")
	require.NotContains(t, audit.Detail, "PASSWORD_SECRET")
}

func TestLDAPConfigAuditOmitsSecrets(t *testing.T) {
	db := openExternalPersistenceDB(t)
	require.NoError(t, db.Create(&systemModel.SystemLDAPConfig{BindPassword: "OLD_SECRET", URL: "ldap://old.test"}).Error)
	svc := systemService.NewLDAPService(&fakeLDAPClientUC{})
	_, err := svc.SaveConfig(context.Background(), 1, systemRequest.LDAPConfigRequest{
		Enabled: true, URL: "ldap://user:URL_SECRET@new.test", BindPassword: "NEW_SECRET",
	}, systemService.AuditMeta{})
	require.NoError(t, err)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "ldap.config.update").First(&audit).Error)
	require.Equal(t, "success", audit.Result)
	require.Contains(t, audit.Detail, "enabled")
	require.Contains(t, audit.Detail, "bindPasswordChanged")
	for _, secret := range []string{"OLD_SECRET", "NEW_SECRET", "URL_SECRET", "ldap://"} {
		require.NotContains(t, audit.Detail, secret)
	}
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
