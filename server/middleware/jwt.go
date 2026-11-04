package middleware

import (
	"strings"

	"optimus/server/model/common/response"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			response.NoAuth("authorization header is required", c)
			c.Abort()
			return
		}

		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(tokenString) == "" {
			response.NoAuth("invalid authorization header", c)
			c.Abort()
			return
		}

		claims, err := utils.NewJWT().ParseToken(strings.TrimSpace(tokenString))
		if err != nil {
			response.NoAuth("invalid or expired token", c)
			c.Abort()
			return
		}

		c.Set(utils.ClaimsContextKey, claims)
		c.Next()
	}
}
