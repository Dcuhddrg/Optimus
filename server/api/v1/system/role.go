package system

import (
	"errors"
	"optimus/server/global"
	"optimus/server/model/common/response"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

type RoleApi struct{}

// swag 只能解析本文件导入的包别名；保留此引用，供 Swagger 注释使用 systemModel。
var _ = systemModel.SystemRole{}

// @Tags Role
// @Summary 角色列表（完整集合，不分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=object{list=[]systemModel.SystemRole,total=int}} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/roles [get]
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

// @Tags Role
// @Summary 创建角色
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body systemRequest.RoleRequest true "请求体"
// @Success 200 {object} response.Response{data=systemModel.SystemRole} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/roles [post]
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

// @Tags Role
// @Summary 更新角色
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id path int true "角色 ID"
// @Param data body systemRequest.RoleRequest true "请求体"
// @Success 200 {object} response.Response{data=systemModel.SystemRole} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/roles/{id} [put]
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

// @Tags Role
// @Summary 启用或禁用角色
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id path int true "角色 ID"
// @Param data body systemRequest.UpdateRoleStatusRequest true "请求体"
// @Success 200 {object} response.Response{data=systemModel.SystemRole} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/roles/{id}/status [patch]
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

// @Tags Role
// @Summary 设置角色权限码
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id path int true "角色 ID"
// @Param data body systemRequest.UpdateRolePermissionsRequest true "请求体"
// @Success 200 {object} response.Response "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/roles/{id}/permissions [put]
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
	if errors.Is(err, systemService.ErrCasbinReloadPending) {
		response.FailWithMessage("permissions saved; reloading, please confirm later", c)
		return
	}
	if err != nil {
		response.FailWithMessage("failed to update role permissions", c)
		return
	}
	response.Ok(c)
}

// @Tags Role
// @Summary 权限码清单（菜单与 API）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=object{list=[]systemModel.SystemPermission}} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/permissions/tree [get]
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
