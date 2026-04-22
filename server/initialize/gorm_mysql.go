package initialize

import (
	"fmt"
	"strings"

	"optimus/server/config"
	"optimus/server/global"

	"go.uber.org/zap"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type sqlDBConn interface {
	SetMaxIdleConns(int)
	SetMaxOpenConns(int)
	Ping() error
	Close() error
}

func GormMysql() *gorm.DB {
	cfg := global.OPTIMUS_CONFIG.Mysql

	db, err := gorm.Open(gormmysql.Open(buildMySQLDSN(cfg)), &gorm.Config{
		Logger: gormlogger.Default.LogMode(mysqlLogMode(cfg.LogMode)),
	})
	if err != nil {
		logDBInitFailure("open mysql connection", err)
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		logDBInitFailure("get mysql sql.DB", err)
		return nil
	}

	if err := configureMySQLPool(sqlDB, cfg); err != nil {
		logDBInitFailure("ping mysql", err)
		return nil
	}

	return db
}

func buildMySQLDSN(cfg config.Mysql) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", cfg.Username, cfg.Password, cfg.Path, cfg.Port, cfg.DBName, cfg.Config)
}

func configureMySQLPool(conn sqlDBConn, cfg config.Mysql) error {
	if cfg.MaxIdleConns > 0 {
		conn.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		conn.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return err
	}

	return nil
}

func mysqlLogMode(mode string) gormlogger.LogLevel {
	switch strings.ToLower(mode) {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "warn", "warning":
		return gormlogger.Warn
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}

func logDBInitFailure(action string, err error) {
	logger, ok := global.OPTIMUS_LOG.(*zap.Logger)
	if !ok || logger == nil {
		return
	}

	logger.Error("mysql init failed", zap.String("action", action), zap.Error(err))
}
