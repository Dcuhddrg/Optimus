package store

import (
	"os"
	"testing"

	cloudModel "optimus/server/model/cloud"
	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/stretchr/testify/require"
)

func openProjectTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec("DROP TABLE IF EXISTS cloud_assets").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS cloud_accounts").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS user_projects").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS projects").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_users").Error)

	require.NoError(t, db.AutoMigrate(
		&systemModel.SystemUser{},
		&projectModel.Project{},
		&projectModel.UserProject{},
		&cloudModel.CloudAccount{},
	))
	return db
}

func seedUser(t *testing.T, db *gorm.DB, id uint, username string) {
	t.Helper()
	u := systemModel.SystemUser{
		ID:       id,
		UUID:     "uuid-" + username,
		Username: username,
		Source:   "local",
		Status:   "active",
	}
	require.NoError(t, db.Create(&u).Error)
}
