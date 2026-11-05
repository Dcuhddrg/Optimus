package task

import (
	"os"
	"testing"
	"time"

	"optimus/server/config"
	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestClearTableDeletesExpiredAuditLogsAndRevokedTokens(t *testing.T) {
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	for _, table := range []string{"sys_audit_logs", "jwt_blacklists"} {
		require.NoError(t, db.Exec("DROP TABLE IF EXISTS `"+table+"`").Error)
	}
	require.NoError(t, systemService.AutoMigrateSystem(db))
	saved := global.OPTIMUS_CONFIG
	t.Cleanup(func() { global.OPTIMUS_CONFIG = saved })
	global.OPTIMUS_CONFIG.System.AuditLogRetentionDays = 30
	global.OPTIMUS_CONFIG.JWT = config.JWT{ExpiresTime: "7d"}

	old, recent := time.Now().AddDate(0, 0, -31), time.Now().AddDate(0, 0, -1)
	require.NoError(t, db.Create(&[]systemModel.SystemAuditLog{
		{Action: "old", Result: "success", CreatedAt: old},
		{Action: "recent", Result: "success", CreatedAt: recent},
	}).Error)
	require.NoError(t, db.Create(&[]systemModel.JwtBlacklist{
		{Jwt: "expired", OPTIMUS_MODEL: global.OPTIMUS_MODEL{CreatedAt: time.Now().AddDate(0, 0, -8)}},
		{Jwt: "live", OPTIMUS_MODEL: global.OPTIMUS_MODEL{CreatedAt: recent}},
	}).Error)

	require.NoError(t, ClearTable(db))

	var actions, tokens []string
	require.NoError(t, db.Model(&systemModel.SystemAuditLog{}).Pluck("action", &actions).Error)
	require.NoError(t, db.Model(&systemModel.JwtBlacklist{}).Pluck("jwt", &tokens).Error)
	require.Equal(t, []string{"recent"}, actions)
	require.Equal(t, []string{"live"}, tokens, "a revoked token stays listed until it would have expired")
}

func TestClearTableRejectsNilDB(t *testing.T) {
	require.Error(t, ClearTable(nil))
}
