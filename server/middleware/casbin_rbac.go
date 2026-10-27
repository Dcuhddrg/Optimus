package middleware

import (
	"context"
	"net/http"
	"strings"

	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type APIAuthorizer interface {
	AuthorizeAPI(ctx context.Context, method string, path string, roleIDs []uint) (bool, error)
}

// CasbinHandlerWithPrefix maps registered route paths to the canonical prefix
// used by permission seeds. It preserves route parameters, methods and roles.
func CasbinHandlerWithPrefix(authorizer APIAuthorizer, runtimePrefix, permissionPrefix string) gin.HandlerFunc {
	if authorizer == nil {
		return CasbinHandler(nil)
	}
	return CasbinHandler(prefixedAPIAuthorizer{
		authorizer:       authorizer,
		runtimePrefix:    strings.TrimRight(runtimePrefix, "/"),
		permissionPrefix: strings.TrimRight(permissionPrefix, "/"),
		valid:            strings.HasPrefix(runtimePrefix, "/") && strings.HasPrefix(permissionPrefix, "/"),
	})
}

type prefixedAPIAuthorizer struct {
	authorizer                      APIAuthorizer
	runtimePrefix, permissionPrefix string
	valid                           bool
}

func (a prefixedAPIAuthorizer) AuthorizeAPI(ctx context.Context, method, path string, roleIDs []uint) (bool, error) {
	if !a.valid || path == "" || !strings.HasPrefix(path, "/") {
		return false, nil
	}
	if a.runtimePrefix != "" && path != a.runtimePrefix && !strings.HasPrefix(path, a.runtimePrefix+"/") {
		return false, nil
	}
	permissionPath := a.permissionPrefix + strings.TrimPrefix(path, a.runtimePrefix)
	return a.authorizer.AuthorizeAPI(ctx, method, permissionPath, roleIDs)
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
