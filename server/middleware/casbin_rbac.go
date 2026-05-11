package middleware

import (
	"context"
	"net/http"

	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type APIAuthorizer interface {
	AuthorizeAPI(ctx context.Context, method string, path string, roleIDs []uint) (bool, error)
}

func CasbinHandler(authorizer APIAuthorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsValue, ok := c.Get(utils.ClaimsContextKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "permission denied"})
			return
		}
		claims, ok := claimsValue.(*utils.CustomClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "permission denied"})
			return
		}

		if authorizer == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "permission denied"})
			return
		}

		allowed, err := authorizer.AuthorizeAPI(c.Request.Context(), c.Request.Method, c.FullPath(), claims.BaseClaims.RoleIDs)
		if err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "permission denied"})
			return
		}

		c.Next()
	}
}
