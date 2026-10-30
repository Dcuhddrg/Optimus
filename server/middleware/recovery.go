package middleware

import (
	"net/http"
	"optimus/server/utils/response"

	"optimus/server/core"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		core.Logger().Error("panic recovered", zap.Any("error", recovered))
		if c.Writer.Written() {
			c.Abort()
			return
		}
		if response.IsEnvelope(c) {
			response.AbortFailure(c, http.StatusInternalServerError, "internal server error")
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	})
}
