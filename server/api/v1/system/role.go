package system

import (
	"errors"
	"net/http"

	"optimus/server/global"
	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

func (a *Api) Roles(c *gin.Context) {
	service, ok := newRoleService(c)
	if !ok {
		return
	}
	roles, err := service.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": roles, "total": len(roles)})
}

func (a *Api) CreateRole(c *gin.Context) {
	service, ok := newRoleService(c)
	if !ok {
		return
	}
	var req systemService.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role request"})
		return
	}
	role, err := service.CreateRole(c.Request.Context(), actorID(c), req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (a *Api) UpdateRole(c *gin.Context) {
	service, ok := newRoleService(c)
	if !ok {
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemService.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role request"})
		return
	}
	role, err := service.UpdateRole(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (a *Api) UpdateRoleStatus(c *gin.Context) {
	service, ok := newRoleService(c)
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
	role, err := service.UpdateRoleStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (a *Api) UpdateRolePermissions(c *gin.Context) {
	service, ok := newRoleService(c)
	if !ok {
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		PermissionCodes []string `json:"permissionCodes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid permissions request"})
		return
	}
	err := service.UpdateRolePermissions(c.Request.Context(), actorID(c), id, req.PermissionCodes, auditMeta(c))
	if errors.Is(err, systemService.ErrRoleNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if errors.Is(err, systemService.ErrProtectedAdminRole) {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update role permissions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role permissions updated"})
}

func (a *Api) PermissionTree(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": systemService.DefaultSystemPermissions()})
}

func newRoleService(c *gin.Context) (*systemService.RoleService, bool) {
	if global.OPTIMUS_DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not available"})
		return nil, false
	}
	return systemService.NewRoleService(systemService.NewGormRoleStore(global.OPTIMUS_DB)), true
}

func writeRoleServiceResult(c *gin.Context, role any, err error) {
	if errors.Is(err, systemService.ErrRoleNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if errors.Is(err, systemService.ErrProtectedAdminRole) {
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, role)
}
