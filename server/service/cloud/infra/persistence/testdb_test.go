package persistence

import (
	"os"
	"testing"

	cloudModel "optimus/server/model/cloud"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/stretchr/testify/require"
)

func openCloudTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec("DROP TABLE IF EXISTS cloud_assets").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS cloud_accounts").Error)
	require.NoError(t, db.AutoMigrate(&cloudModel.CloudAccount{}, &cloudModel.CloudAsset{}))

	return db
}
