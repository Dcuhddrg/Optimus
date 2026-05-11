package system

import (
	"errors"
	"net/http"
	"strconv"

	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (h *UserHandler) Users(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "user service is not available")
		return
	}
	result, err := h.service.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "user service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemService.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user request"})
		return
	}
	user, err := h.service.UpdateUser(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "user service is not available")
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid status request"})
		return
	}
	user, err := h.service.UpdateUserStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

func (h *UserHandler) UpdateUserRoles(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "user service is not available")
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
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role request"})
		return
	}
	err := h.service.UpdateUserRoles(c.Request.Context(), actorID(c), id, req.RoleIDs, auditMeta(c))
	if errors.Is(err, systemService.ErrLastAdminCapableUser) {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user roles updated"})
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
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
	return systemService.AuditMeta{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
}

func writeUserServiceResult(c *gin.Context, user systemService.SystemUserSummary, err error) {
	if errors.Is(err, systemService.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if errors.Is(err, systemService.ErrLastAdminCapableUser) {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update user"})
		return
	}
	c.JSON(http.StatusOK, user)
}
