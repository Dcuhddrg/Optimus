package project

import (
	"errors"
	"net/http"
	"strconv"

	projectSvc "optimus/server/service/project"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func writeProjectServiceUnavailable(c *gin.Context) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": "project service is not available"})
}

func writeProjectError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, projectSvc.ErrProjectNotFound),
		errors.Is(err, projectSvc.ErrMemberNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, projectSvc.ErrProjectCodeTaken),
		errors.Is(err, projectSvc.ErrProjectHasAccounts),
		errors.Is(err, projectSvc.ErrCannotRemoveLastOwner):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, projectSvc.ErrInvalidProjectCode),
		errors.Is(err, projectSvc.ErrInvalidProjectName),
		errors.Is(err, projectSvc.ErrInvalidProjectRole),
		errors.Is(err, projectSvc.ErrUserNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, projectSvc.ErrForbidden):
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
