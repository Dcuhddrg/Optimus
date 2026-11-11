package system

import (
	"os"
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	"optimus/server/utils"
	ldapUtil "optimus/server/utils/ldap"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// resetTestDB opens OPTIMUS_TEST_MYSQL_DSN and drops the system tables. The DSN
// must point to a database reserved for tests.
func resetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	for _, table := range []string{"sys_role_permissions", "sys_user_roles", "sys_permissions", "sys_roles", "sys_users", "sys_ldap_config", "sys_audit_logs", "jwt_blacklists", "casbin_rule"} {
		require.NoError(t, db.Exec("DROP TABLE IF EXISTS `"+table+"`").Error)
	}
	return db
}

// openTestDB resets and migrates the system tables and installs the database
// as global.OPTIMUS_DB until the test ends.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := resetTestDB(t)
	require.NoError(t, AutoMigrateSystem(db))
	previous := global.OPTIMUS_DB
	t.Cleanup(func() { global.OPTIMUS_DB = previous; utils.ResetCasbin() })
	global.OPTIMUS_DB = db
	utils.ResetCasbin()
	InvalidateAllAuthState()
	t.Cleanup(InvalidateAllAuthState)
	return db
}

// newLDAPServiceForTest makes LDAP calls use client until the test ends.
func newLDAPServiceForTest(t testing.TB, client ldapUtil.Client) *LDAPService {
	t.Helper()
	saved := newLDAPClient
	t.Cleanup(func() { newLDAPClient = saved })
	newLDAPClient = func() ldapUtil.Client { return client }
	return &LDAPService{}
}

// newAuthServiceForTest sets a signing config so login can issue tokens.
func newAuthServiceForTest(t testing.TB) *AuthService {
	t.Helper()
	saved := global.OPTIMUS_CONFIG.JWT
	t.Cleanup(func() { global.OPTIMUS_CONFIG.JWT = saved })
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "test-signing-key", ExpiresTime: "1h", Issuer: "optimus-test"}
	return &AuthService{}
}
