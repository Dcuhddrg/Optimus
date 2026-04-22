package initialize

import (
	"strings"

	"optimus/server/global"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	var db *gorm.DB

	switch strings.ToLower(global.OPTIMUS_CONFIG.System.DBType) {
	case "mysql":
		db = GormMysql()
	default:
		logUnsupportedDBType(global.OPTIMUS_CONFIG.System.DBType)
	}

	global.OPTIMUS_DB = db
	return db
}

func logUnsupportedDBType(dbType string) {
	logger, ok := global.OPTIMUS_LOG.(*zap.Logger)
	if !ok || logger == nil {
		return
	}

	logger.Error("unsupported db type", zap.String("db_type", dbType))
}
