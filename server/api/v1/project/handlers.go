package project

import (
	"errors"
	"net/http"
	"strconv"

	projectUsecase "optimus/server/service/project/usecase"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type ProjectHandlers struct {
	svc *projectUsecase.ProjectService
}

type ProjectHandlerDeps struct {
	Service *projectUsecase.ProjectService
}

func NewProjectHandlers(deps ProjectHandlerDeps) *ProjectHandlers {
	return &ProjectHandlers{svc: deps.Service}
}

func writeProjectServiceUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": "project service is not available"})
}

func writeProjectError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, projectUsecase.ErrProjectNotFound),
		errors.Is(err, projectUsecase.ErrMemberNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, projectUsecase.ErrProjectCodeTaken),
		errors.Is(err, projectUsecase.ErrProjectHasAccounts),
		errors.Is(err, projectUsecase.ErrCannotRemoveLastOwner):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, projectUsecase.ErrInvalidProjectCode),
		errors.Is(err, projectUsecase.ErrInvalidProjectName),
		errors.Is(err, projectUsecase.ErrInvalidProjectRole),
		errors.Is(err, projectUsecase.ErrUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, projectUsecase.ErrForbidden):
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
