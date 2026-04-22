package core

import (
	"os"
	"strings"

	"optimus/server/global"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Zap() error {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"

	var encoder zapcore.Encoder
	switch strings.ToLower(global.OPTIMUS_CONFIG.Zap.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stdout),
		zapLevel(global.OPTIMUS_CONFIG.Zap.Level),
	)

	options := []zap.Option{zap.AddCaller(), zap.ErrorOutput(zapcore.Lock(os.Stderr))}
	if prefix := strings.TrimSpace(global.OPTIMUS_CONFIG.Zap.Prefix); prefix != "" {
		options = append(options, zap.Fields(zap.String("service", prefix)))
	}

	logger := zap.New(core, options...)
	global.OPTIMUS_LOG = logger
	zap.ReplaceGlobals(logger)
	return nil
}

func Logger() *zap.Logger {
	logger, ok := global.OPTIMUS_LOG.(*zap.Logger)
	if ok && logger != nil {
		return logger
	}

	return zap.L()
}

func zapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
