package system

import (
	"errors"
	"net/http"
	"strconv"

	"optimus/server/global"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func (a *Api) Users(c *gin.Context) {
	service, ok := newUserService(c)
	if !ok {
		return
	}
	result, err := service.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (a *Api) UpdateUser(c *gin.Context) {
	service, ok := newUserService(c)
	if !ok {
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
	user, err := service.UpdateUser(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

func (a *Api) UpdateUserStatus(c *gin.Context) {
	service, ok := newUserService(c)
	if !ok {
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
	user, err := service.UpdateUserStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

func (a *Api) UpdateUserRoles(c *gin.Context) {
	service, ok := newUserService(c)
	if !ok {
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
	err := service.UpdateUserRoles(c.Request.Context(), actorID(c), id, req.RoleIDs, auditMeta(c))
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

func newUserService(c *gin.Context) (*systemService.UserService, bool) {
	if global.OPTIMUS_DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not available"})
		return nil, false
	}
	return systemService.NewUserService(systemService.NewGormUserStore(global.OPTIMUS_DB)), true
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
