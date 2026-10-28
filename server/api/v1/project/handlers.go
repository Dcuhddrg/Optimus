package project

import (
	"errors"
	"net/http"
	"strconv"

	projectService "optimus/server/service/project"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type ProjectHandlers struct {
	svc *projectService.ProjectService
}

type ProjectHandlerDeps struct {
	Service *projectService.ProjectService
}

func NewProjectHandlers(deps ProjectHandlerDeps) *ProjectHandlers {
	return &ProjectHandlers{svc: deps.Service}
}

func writeProjectServiceUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": "project service is not available"})
}

func writeProjectError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, projectService.ErrProjectNotFound),
		errors.Is(err, projectService.ErrMemberNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, projectService.ErrProjectCodeTaken),
		errors.Is(err, projectService.ErrProjectHasAccounts),
		errors.Is(err, projectService.ErrCannotRemoveLastOwner):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, projectService.ErrInvalidProjectCode),
		errors.Is(err, projectService.ErrInvalidProjectName),
		errors.Is(err, projectService.ErrInvalidProjectRole),
		errors.Is(err, projectService.ErrUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, projectService.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
	}
}

func projectActorID(c *gin.Context) uint {
	v, exists := c.Get(utils.ClaimsContextKey)
	if !exists {
		return 0
	}
	claims, ok := v.(*utils.CustomClaims)
	if !ok {
		return 0
	}
	return claims.BaseClaims.ID
}

func projectParseUintParam(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || n == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid " + name})
		return 0, false
	}
	return uint(n), true
}
