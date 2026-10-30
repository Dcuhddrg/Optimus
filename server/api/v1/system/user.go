package system

import (
	"errors"
	"net/http"
	systemModel "optimus/server/model/system"
	"strconv"

	systemService "optimus/server/service/system"
	"optimus/server/utils"
	"optimus/server/utils/response"

	"github.com/gin-gonic/gin"
)

func (h *UserHandler) Users(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "user service is not available")
		return
	}
	result, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to list users")
		return
	}
	response.Success(c, http.StatusOK, gin.H{"list": result.Items, "total": result.Total})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "user service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemModel.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid user request")
		return
	}
	user, err := h.service.UpdateUser(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "user service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid status request")
		return
	}
	user, err := h.service.UpdateUserStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

func (h *UserHandler) UpdateUserRoles(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "user service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		RoleIDs []uint `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Failure(c, http.StatusBadRequest, "invalid role request")
		return
	}
	err := h.service.UpdateUserRoles(c.Request.Context(), actorID(c), id, req.RoleIDs, auditMeta(c))
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

func writeUserServiceResult(c *gin.Context, user systemModel.SystemUserSummary, err error) {
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
