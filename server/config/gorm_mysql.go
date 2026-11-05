package config

import (
	"strings"

	gormlogger "gorm.io/gorm/logger"
)

type Mysql struct {
	Path         string `mapstructure:"path"`
	Port         int    `mapstructure:"port"`
	DBName       string `mapstructure:"db-name"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Config       string `mapstructure:"config"`
	MaxIdleConns int    `mapstructure:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns"`
	LogMode      string `mapstructure:"log-mode"`
	// As GVA's GeneralDB: forward GORM logs to zap, table prefix and singular names.
	LogZap   bool   `mapstructure:"log-zap"`
	Prefix   string `mapstructure:"prefix"`
	Singular bool   `mapstructure:"singular"`
}

// LogLevel maps log-mode to the GORM log level, as GVA's GeneralDB.LogLevel.
func (m Mysql) LogLevel() gormlogger.LogLevel {
	switch strings.ToLower(m.LogMode) {
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
