package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger().Error("panic recovered", zap.Any("error", recovered))
		if c.Writer.Written() {
			c.Abort()
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
