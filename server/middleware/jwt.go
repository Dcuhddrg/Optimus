package middleware

import (
	"net/http"
	"strings"

	"optimus/server/global"
	systemService "optimus/server/service/system"
	"optimus/server/utils"
	"optimus/server/utils/response"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			response.AbortFailure(c, http.StatusUnauthorized, "authorization header is required")
			return
		}

		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(tokenString) == "" {
			response.AbortFailure(c, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		claims, err := systemService.ParseJWT(strings.TrimSpace(tokenString), global.OPTIMUS_CONFIG.JWT)
		if err != nil {
			response.AbortFailure(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		c.Set(utils.ClaimsContextKey, claims)
		c.Next()
	}
}
