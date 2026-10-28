package system

import (
	"errors"
	"net/http"
	systemModel "optimus/server/model/system"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

func (h *RoleHandler) Roles(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	roles, err := h.service.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": roles, "total": len(roles)})
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	var req systemModel.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role request"})
		return
	}
	role, err := h.service.CreateRole(c.Request.Context(), actorID(c), req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemModel.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid role request"})
		return
	}
	role, err := h.service.UpdateRole(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleHandler) UpdateRoleStatus(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "role service is not available")
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
	role, err := h.service.UpdateRoleStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleHandler) UpdateRolePermissions(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "role service is not available")
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
	err := h.service.UpdateRolePermissions(c.Request.Context(), actorID(c), id, req.PermissionCodes, auditMeta(c))
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

func (h *RoleHandler) PermissionTree(c *gin.Context) {
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
