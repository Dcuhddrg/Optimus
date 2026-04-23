package middleware

import "github.com/gin-gonic/gin"

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Auth enforcement is added later; the skeleton only preserves the hook point.
		c.Next()
	}
}
