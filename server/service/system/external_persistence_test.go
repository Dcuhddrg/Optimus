package system_test

import (
	"os"
	"testing"

	"optimus/server/global"
	service "optimus/server/service/system"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// The DSN must point to a database reserved for tests; its system tables are reset.
func openExternalPersistenceDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	for _, table := range []string{"sys_role_permissions", "sys_user_roles", "sys_permissions", "sys_roles", "sys_users", "sys_ldap_config", "sys_audit_logs"} {
		require.NoError(t, db.Exec("DROP TABLE IF EXISTS `"+table+"`").Error)
	}
	require.NoError(t, service.AutoMigrateSystem(db))
	previous := global.OPTIMUS_DB
	global.OPTIMUS_DB = db
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { global.OPTIMUS_DB = previous; require.NoError(t, sqlDB.Close()) })
	return db
}
