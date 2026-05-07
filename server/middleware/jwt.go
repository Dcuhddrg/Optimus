package middleware

import (
	"net/http"
	"strings"

	"optimus/server/global"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "authorization header is required"})
			return
		}

		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(tokenString) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization header"})
			return
		}

		claims, err := systemService.ParseJWT(strings.TrimSpace(tokenString), global.OPTIMUS_CONFIG.JWT)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired token"})
			return
		}

		c.Set(utils.ClaimsContextKey, claims)
		c.Next()
	}
}
