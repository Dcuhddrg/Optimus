package project

import (
	"errors"
	"strconv"

	"optimus/server/model/common/response"
	projectSvc "optimus/server/service/project"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func writeProjectServiceUnavailable(c *gin.Context) {
	response.FailWithMessage("project service is not available", c)
}

func writeProjectError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, projectSvc.ErrProjectNotFound),
		errors.Is(err, projectSvc.ErrMemberNotFound):
		response.FailWithMessage(err.Error(), c)
	case errors.Is(err, projectSvc.ErrProjectCodeTaken),
		errors.Is(err, projectSvc.ErrProjectHasAccounts),
		errors.Is(err, projectSvc.ErrCannotRemoveLastOwner):
		response.FailWithMessage(err.Error(), c)
	case errors.Is(err, projectSvc.ErrInvalidProjectCode),
		errors.Is(err, projectSvc.ErrInvalidProjectName),
		errors.Is(err, projectSvc.ErrInvalidProjectRole),
		errors.Is(err, projectSvc.ErrUserNotFound),
		errors.Is(err, projectSvc.ErrExternalCannotOwn):
		response.FailWithMessage(err.Error(), c)
	case errors.Is(err, projectSvc.ErrForbidden):
		response.FailWithMessage(err.Error(), c)
	default:
		response.FailWithMessage("internal error", c)
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
		response.FailWithMessage("invalid "+name, c)
		return 0, false
	}
	return uint(n), true
}
