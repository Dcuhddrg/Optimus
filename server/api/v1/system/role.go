package system

import (
	"errors"
	"net/http"
	systemRequest "optimus/server/model/system/request"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

type RoleApi struct{}

func (h *RoleApi) Roles(c *gin.Context) {
	svc := *roleService
	if svc == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	roles, err := svc.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": roles, "total": len(roles)})
}

func (h *RoleApi) CreateRole(c *gin.Context) {
	svc := *roleService
	if svc == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	var req systemRequest.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role request"})
		return
	}
	role, err := svc.CreateRole(c.Request.Context(), actorID(c), req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleApi) UpdateRole(c *gin.Context) {
	svc := *roleService
	if svc == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role request"})
		return
	}
	role, err := svc.UpdateRole(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleApi) UpdateRoleStatus(c *gin.Context) {
	svc := *roleService
	if svc == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateRoleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid status request"})
		return
	}
	role, err := svc.UpdateRoleStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleApi) UpdateRolePermissions(c *gin.Context) {
	svc := *roleService
	if svc == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid permissions request"})
		return
	}
	err := svc.UpdateRolePermissions(c.Request.Context(), actorID(c), id, req.PermissionCodes, auditMeta(c))
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

func (h *RoleApi) PermissionTree(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": systemService.DefaultSystemPermissions()})
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
