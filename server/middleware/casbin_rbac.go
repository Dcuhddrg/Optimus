package middleware

import (
	"context"
	"net/http"

	"optimus/server/global"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type rbacAuthorizer interface {
	AuthorizeAPI(ctx context.Context, method string, path string, roleIDs []uint) (bool, error)
}

var newRBACAuthorizer = func() rbacAuthorizer {
	if global.OPTIMUS_DB == nil {
		return nil
	}
	return systemService.NewRBACService(systemService.NewGormRBACStore(global.OPTIMUS_DB))
}

func CasbinHandler() gin.HandlerFunc {
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

		authorizer := newRBACAuthorizer()
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
