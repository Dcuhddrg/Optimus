package middleware

import (
	"context"
	"net/http"

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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to resolve project scope"})
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to resolve project"})
			return
		}
		if !exists {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "project not found"})
			return
		}
		if !containsUint(accessible, projectID) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "no access to project"})
			return
		}

		role, exists, err := resolver.GetMemberRole(c.Request.Context(), claims.BaseClaims.ID, projectID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to resolve project role"})
			return
		}
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "no access to project"})
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
