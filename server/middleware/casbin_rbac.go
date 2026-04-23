package middleware

import "github.com/gin-gonic/gin"

func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// RBAC enforcement is added later; the skeleton only preserves the hook point.
		c.Next()
	}
}
