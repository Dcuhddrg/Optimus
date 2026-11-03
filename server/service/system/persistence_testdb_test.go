package system

import (
	"os"
	"testing"

	"optimus/server/global"

	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func openPersistenceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}

	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_role_permissions").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_user_roles").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_permissions").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_roles").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_users").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_ldap_config").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_audit_logs").Error)

	require.NoError(t, AutoMigrateSystem(db))

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })

	previousDB := global.OPTIMUS_DB
	t.Cleanup(func() { global.OPTIMUS_DB = previousDB })
	global.OPTIMUS_DB = db
	return db
}
