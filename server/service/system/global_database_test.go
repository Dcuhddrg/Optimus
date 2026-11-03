package system

import (
	"context"
	"testing"

	"optimus/server/global"

	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRoleServiceQueriesStartupDatabase(t *testing.T) {
	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       "unused:unused@tcp(127.0.0.1:3306)/unused",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	oldDB := global.OPTIMUS_DB
	t.Cleanup(func() { global.OPTIMUS_DB = oldDB })
	global.OPTIMUS_DB = db

	var query string
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("test:capture", func(tx *gorm.DB) {
		query = tx.Statement.SQL.String()
	}))

	require.NotPanics(t, func() {
		_, err := (&RoleService{}).ListRoles(context.Background())
		require.NoError(t, err)
	})
	require.Contains(t, query, "sys_roles")
	require.Contains(t, query, "ORDER BY id ASC")
}
