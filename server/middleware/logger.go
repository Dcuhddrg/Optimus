package middleware

import (
	"time"

	"optimus/server/global"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger().Info(
			"http request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func logger() *zap.Logger {
	if logger, ok := global.OPTIMUS_LOG.(*zap.Logger); ok && logger != nil {
		return logger
	}
	return zap.L()
}
