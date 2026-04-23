package middleware

import (
	"net/http"

	"optimus/server/core"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		core.Logger().Error("panic recovered", zap.Any("error", recovered))
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
