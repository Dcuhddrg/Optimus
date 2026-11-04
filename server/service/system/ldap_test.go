package system

import (
	"context"
	"errors"
	"testing"

	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestLDAPServiceConfigMasksBindPassword(t *testing.T) {
	db := openTestDB(t)
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
	service := newLDAPServiceForTest(t, &fakeLDAPClientUC{})

	config, err := service.GetConfig(context.Background())

	require.NoError(t, err)
	require.True(t, config.HasBindPassword)
	require.Empty(t, config.BindPassword)
}

type failingLDAPClient struct{}

func (failingLDAPClient) Test(context.Context, LDAPConnectionConfig) error {
	return errors.New("DO_NOT_LOG_THIS")
}

func (failingLDAPClient) Authenticate(context.Context, LDAPConnectionConfig, string, string) (LDAPUser, error) {
	return LDAPUser{}, errors.New("DO_NOT_LOG_THIS")
}

func TestLDAPSnapshotFailureStillSavesConfig(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("test:fail_ldap_read", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_ldap_config" {
			tx.AddError(errors.New("snapshot read failed"))
		}
	}))
	svc := newLDAPServiceForTest(t, &fakeLDAPClientUC{})
	_, err := svc.SaveConfig(context.Background(), 1, systemRequest.LDAPConfigRequest{Enabled: true, BindPassword: "SECRET"}, AuditMeta{})
	require.NoError(t, err)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "success", audit.Result)
	require.Contains(t, audit.Detail, "snapshot_unavailable")
	require.NotContains(t, audit.Detail, "SECRET")
}

func TestLDAPConnectionFailureAuditOmitsSecrets(t *testing.T) {
	db := openTestDB(t)
	svc := newLDAPServiceForTest(t, failingLDAPClient{})
	_, err := svc.TestConfig(context.Background(), 1, systemRequest.LDAPTestRequest{
		URL: "ldap://test", BindPassword: "PASSWORD_SECRET", UserFilter: "(uid={username})",
	}, AuditMeta{})
	require.Error(t, err)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "ldap.test.failure").First(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "connection_test_failed")
	require.NotContains(t, audit.Detail, "DO_NOT_LOG_THIS")
	require.NotContains(t, audit.Detail, "PASSWORD_SECRET")
}

func TestLDAPConfigAuditOmitsSecrets(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Create(&systemModel.SystemLDAPConfig{BindPassword: "OLD_SECRET", URL: "ldap://old.test"}).Error)
	svc := newLDAPServiceForTest(t, &fakeLDAPClientUC{})
	_, err := svc.SaveConfig(context.Background(), 1, systemRequest.LDAPConfigRequest{
		Enabled: true, URL: "ldap://user:URL_SECRET@new.test", BindPassword: "NEW_SECRET",
	}, AuditMeta{})
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

type fakeLDAPClientUC struct{}

func (c *fakeLDAPClientUC) Test(_ context.Context, cfg LDAPConnectionConfig) error {
	return nil
}

func (c *fakeLDAPClientUC) Authenticate(context.Context, LDAPConnectionConfig, string, string) (LDAPUser, error) {
	return LDAPUser{}, errors.New("not used")
}

func TestLDAPServiceDatabaseGetLDAPConfigReturnsEmptyWhenMissing(t *testing.T) {
	openTestDB(t)

	store := (&LDAPService{})
	cfg, err := store.getLDAPConfig(context.Background())

	require.NoError(t, err)
	require.Equal(t, uint(0), cfg.ID)
}
