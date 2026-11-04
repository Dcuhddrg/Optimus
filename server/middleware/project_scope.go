package middleware

import (
	"context"

	"optimus/server/model/common/response"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type ProjectScopeResolver interface {
	ListAccessibleProjectIDs(ctx context.Context, userID uint) ([]uint, error)
	ResolveProjectByCode(ctx context.Context, code string) (projectID uint, exists bool, err error)
	GetMemberRole(ctx context.Context, userID, projectID uint) (role string, exists bool, err error)
}

func ProjectScope(resolver ProjectScopeResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		if resolver == nil {
			c.Next()
			return
		}

		rawClaims, exists := c.Get(utils.ClaimsContextKey)
		if !exists {
			c.Next()
			return
		}
		claims, ok := rawClaims.(*utils.CustomClaims)
		if !ok || claims == nil {
			c.Next()
			return
		}

		accessible, err := resolver.ListAccessibleProjectIDs(c.Request.Context(), claims.BaseClaims.ID)
		if err != nil {
			response.FailWithMessage("failed to resolve project scope", c)
			c.Abort()
			return
		}

		scope := utils.ProjectScope{AccessibleProjectIDs: accessible}
		code := c.GetHeader("X-Project-Code")
		if code == "" {
			c.Set(utils.ProjectScopeContextKey, scope)
			c.Next()
			return
		}

		projectID, exists, err := resolver.ResolveProjectByCode(c.Request.Context(), code)
		if err != nil {
			response.FailWithMessage("failed to resolve project", c)
			c.Abort()
			return
		}
		if !exists {
			response.FailWithMessage("project not found", c)
			c.Abort()
			return
		}
		if !containsUint(accessible, projectID) {
			response.FailWithMessage("no access to project", c)
			c.Abort()
			return
		}

		role, exists, err := resolver.GetMemberRole(c.Request.Context(), claims.BaseClaims.ID, projectID)
		if err != nil {
			response.FailWithMessage("failed to resolve project role", c)
			c.Abort()
			return
		}
		if !exists {
			response.FailWithMessage("no access to project", c)
			c.Abort()
			return
		}

		scope.CurrentProjectID = projectID
		scope.CurrentProjectCode = code
		scope.CurrentProjectRole = role
		c.Set(utils.ProjectScopeContextKey, scope)
		c.Next()
	}
}

func containsUint(values []uint, target uint) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
