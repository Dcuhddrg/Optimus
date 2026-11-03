package system

import (
	"errors"
	"net/http"
	"optimus/server/global"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"
	"strconv"

	"optimus/server/model/common/response"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type UserApi struct{}

func (h *UserApi) Users(c *gin.Context) {
	svc := userService
	if global.OPTIMUS_DB == nil {
		response.Failure(c, http.StatusServiceUnavailable, "user service is not available")
		return
	}
	result, err := svc.ListUsers(c.Request.Context())
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to list users")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"list": result.Items, "total": result.Total})
}

func (h *UserApi) UpdateUser(c *gin.Context) {
	svc := userService
	if global.OPTIMUS_DB == nil {
		response.Failure(c, http.StatusServiceUnavailable, "user service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid user request")
		return
	}
	user, err := svc.UpdateUser(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

func (h *UserApi) UpdateUserStatus(c *gin.Context) {
	svc := userService
	if global.OPTIMUS_DB == nil {
		response.Failure(c, http.StatusServiceUnavailable, "user service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid status request")
		return
	}
	user, err := svc.UpdateUserStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

func (h *UserApi) UpdateUserRoles(c *gin.Context) {
	svc := userService
	if global.OPTIMUS_DB == nil {
		response.Failure(c, http.StatusServiceUnavailable, "user service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid role request")
		return
	}
	err := svc.UpdateUserRoles(c.Request.Context(), actorID(c), id, req.RoleIDs, auditMeta(c))
	if errors.Is(err, systemService.ErrLastAdminCapableUser) {
		response.Failure(c, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to update user roles")
		return
	}
	response.Success(c, http.StatusOK, nil)
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		if response.IsEnvelope(c) {
			response.Failure(c, http.StatusBadRequest, "invalid id")
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		}
		return 0, false
	}
	return uint(value), true
}

func actorID(c *gin.Context) uint {
	claims, ok := c.Get(utils.ClaimsContextKey)
	if !ok {
		return 0
	}
	customClaims, ok := claims.(*utils.CustomClaims)
	if !ok {
		return 0
	}
	return customClaims.BaseClaims.ID
}

func auditMeta(c *gin.Context) systemService.AuditMeta {
	meta := systemService.AuditMeta{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
	if raw, ok := c.Get(utils.ClaimsContextKey); ok {
		if cc, ok := raw.(*utils.CustomClaims); ok {
			meta.ActorUsername = cc.BaseClaims.Username
		}
	}
	return meta
}

func writeUserServiceResult(c *gin.Context, user systemResponse.SystemUserSummary, err error) {
	if errors.Is(err, systemService.ErrUserNotFound) {
		response.Failure(c, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, systemService.ErrLastAdminCapableUser) {
		response.Failure(c, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to update user")
		return
	}
	response.Success(c, http.StatusOK, user)
}
