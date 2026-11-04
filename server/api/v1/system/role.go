package system

import (
	"errors"
	"optimus/server/global"
	"optimus/server/model/common/response"
	systemRequest "optimus/server/model/system/request"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

type RoleApi struct{}

func (h *RoleApi) Roles(c *gin.Context) {
	svc := roleService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	roles, err := svc.ListRoles(c.Request.Context())
	if err != nil {
		response.FailWithMessage("failed to list roles", c)
		return
	}
	response.OkWithData(gin.H{"list": roles, "total": len(roles)}, c)
}

func (h *RoleApi) CreateRole(c *gin.Context) {
	svc := roleService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	var req systemRequest.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid role request", c)
		return
	}
	role, err := svc.CreateRole(c.Request.Context(), actorID(c), req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleApi) UpdateRole(c *gin.Context) {
	svc := roleService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid role request", c)
		return
	}
	role, err := svc.UpdateRole(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleApi) UpdateRoleStatus(c *gin.Context) {
	svc := roleService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateRoleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid status request", c)
		return
	}
	role, err := svc.UpdateRoleStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeRoleServiceResult(c, role, err)
}

func (h *RoleApi) UpdateRolePermissions(c *gin.Context) {
	svc := roleService
	if global.OPTIMUS_DB == nil {
		writeServiceUnavailable(c, "role service is not available")
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid permissions request", c)
		return
	}
	err := svc.UpdateRolePermissions(c.Request.Context(), actorID(c), id, req.PermissionCodes, auditMeta(c))
	if errors.Is(err, systemService.ErrRoleNotFound) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if errors.Is(err, systemService.ErrProtectedAdminRole) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err != nil {
		response.FailWithMessage("failed to update role permissions", c)
		return
	}
	response.Ok(c)
}

func (h *RoleApi) PermissionTree(c *gin.Context) {
	response.OkWithData(gin.H{"list": systemService.DefaultSystemPermissions()}, c)
}

func writeRoleServiceResult(c *gin.Context, role any, err error) {
	if errors.Is(err, systemService.ErrRoleNotFound) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if errors.Is(err, systemService.ErrProtectedAdminRole) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err != nil {
		response.FailWithMessage("failed to update role", c)
		return
	}
	response.OkWithData(role, c)
}
