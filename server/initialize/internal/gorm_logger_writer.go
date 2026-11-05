package internal

import (
	"fmt"

	"optimus/server/config"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// Writer prints GORM logs to stdout and, with log-zap, to zap, as GVA.
type Writer struct {
	config config.Mysql
}

func NewWriter(config config.Mysql) *Writer {
	return &Writer{config: config}
}

func (c *Writer) Printf(message string, data ...any) {
	fmt.Printf(message, data...)
	if !c.config.LogZap {
		return
	}
	switch c.config.LogLevel() {
	case logger.Silent:
		zap.L().Debug(fmt.Sprintf(message, data...))
	case logger.Error:
		zap.L().Error(fmt.Sprintf(message, data...))
	case logger.Warn:
		zap.L().Warn(fmt.Sprintf(message, data...))
	case logger.Info:
		zap.L().Info(fmt.Sprintf(message, data...))
	default:
		zap.L().Info(fmt.Sprintf(message, data...))
	}
}
