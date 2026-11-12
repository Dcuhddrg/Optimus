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

func (failingLDAPClient) LookupUsername(context.Context, LDAPConnectionConfig, string) (LDAPUser, error) {
	return LDAPUser{}, errors.New("DO_NOT_LOG_THIS")
}

func (failingLDAPClient) SearchScope(context.Context, LDAPConnectionConfig) ([]LDAPUser, error) {
	return nil, errors.New("DO_NOT_LOG_THIS")
}

func (failingLDAPClient) LookupExternalID(context.Context, LDAPConnectionConfig, string) (LDAPUser, error) {
	return LDAPUser{}, errors.New("DO_NOT_LOG_THIS")
}

func (failingLDAPClient) ExistsInDirectory(context.Context, LDAPConnectionConfig, string) (bool, error) {
	return false, errors.New("DO_NOT_LOG_THIS")
}

// A failed read of the current config refuses the save even with no bound
// users: building on an empty snapshot would insert a second row that readers
// never pick, and report success while the old config stays in effect.
func TestSaveLDAPConfigRefusesWhenCurrentConfigUnreadable(t *testing.T) {
	db := openTestDB(t)
	svc := newLDAPServiceForTest(t, newFakeDirectory())
	first := validLDAPRequest()
	first.BindPassword = "SECRET"
	_, err := svc.SaveConfig(context.Background(), 1, first, AuditMeta{})
	require.NoError(t, err)

	failNext := true
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("test:fail_ldap_read", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_ldap_config" && failNext {
			failNext = false
			tx.AddError(errors.New("transient read failure"))
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Query().Remove("test:fail_ldap_read") })

	changed := validLDAPRequest()
	changed.BaseDN = "ou=changed,dc=example,dc=com"
	_, err = svc.SaveConfig(context.Background(), 1, changed, AuditMeta{})
	require.Error(t, err)

	var rows int64
	require.NoError(t, db.Model(&systemModel.SystemLDAPConfig{}).Count(&rows).Error)
	require.EqualValues(t, 1, rows, "no second row is inserted")
	cfg, err := svc.getLDAPConfig(context.Background())
	require.NoError(t, err)
	require.Equal(t, first.BaseDN, cfg.BaseDN)
	require.EqualValues(t, 1, cfg.Version)

	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "ldap.config.update").Last(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "update_failed")
	require.NotContains(t, audit.Detail, "SECRET")
}

func TestLDAPConnectionFailureAuditOmitsSecrets(t *testing.T) {
	db := openTestDB(t)
	svc := newLDAPServiceForTest(t, failingLDAPClient{})
	_, err := svc.TestConfig(context.Background(), 1, systemRequest.LDAPTestRequest{
		URL: "ldap://test", BindPassword: "PASSWORD_SECRET", UserFilter: "(uid={username})", UsernameAttr: "uid",
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
		Enabled: true, URL: "ldap://user:URL_SECRET@new.test", BindPassword: "NEW_SECRET", UserFilter: "(uid={username})", UsernameAttr: "uid",
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

func (c *fakeLDAPClientUC) LookupUsername(context.Context, LDAPConnectionConfig, string) (LDAPUser, error) {
	return LDAPUser{}, errors.New("not used")
}

func (c *fakeLDAPClientUC) SearchScope(context.Context, LDAPConnectionConfig) ([]LDAPUser, error) {
	return nil, errors.New("not used")
}

func (c *fakeLDAPClientUC) LookupExternalID(context.Context, LDAPConnectionConfig, string) (LDAPUser, error) {
	return LDAPUser{}, errors.New("not used")
}

func (c *fakeLDAPClientUC) ExistsInDirectory(context.Context, LDAPConnectionConfig, string) (bool, error) {
	return false, errors.New("not used")
}

func TestLDAPServiceDatabaseGetLDAPConfigReturnsEmptyWhenMissing(t *testing.T) {
	openTestDB(t)

	store := (&LDAPService{})
	cfg, err := store.getLDAPConfig(context.Background())

	require.NoError(t, err)
	require.Equal(t, uint(0), cfg.ID)
}
