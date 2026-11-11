package system

import (
	"errors"
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

// @Tags User
// @Summary 用户列表（完整集合，不分页）
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=object{list=[]systemResponse.UserListItem,total=int}} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/users [get]
func (h *UserApi) Users(c *gin.Context) {
	svc := userService
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("user service is not available", c)
		return
	}
	result, err := svc.ListUsers(c.Request.Context())
	if err != nil {
		response.FailWithMessage("failed to list users", c)
		return
	}
	response.OkWithData(gin.H{"list": result.Items, "total": result.Total}, c)
}

// @Tags User
// @Summary 更新用户资料
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id path int true "用户 ID"
// @Param data body systemRequest.UpdateUserRequest true "请求体"
// @Success 200 {object} response.Response{data=systemResponse.SystemUserSummary} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/users/{id} [put]
func (h *UserApi) UpdateUser(c *gin.Context) {
	svc := userService
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("user service is not available", c)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid user request", c)
		return
	}
	user, err := svc.UpdateUser(c.Request.Context(), actorID(c), id, req, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

// @Tags User
// @Summary 启用或禁用用户
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id path int true "用户 ID"
// @Param data body systemRequest.UpdateUserStatusRequest true "请求体"
// @Success 200 {object} response.Response{data=systemResponse.SystemUserSummary} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/users/{id}/status [patch]
func (h *UserApi) UpdateUserStatus(c *gin.Context) {
	svc := userService
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("user service is not available", c)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid status request", c)
		return
	}
	user, err := svc.UpdateUserStatus(c.Request.Context(), actorID(c), id, req.Status, auditMeta(c))
	writeUserServiceResult(c, user, err)
}

// @Tags User
// @Summary 设置用户角色
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id path int true "用户 ID"
// @Param data body systemRequest.UpdateUserRolesRequest true "请求体"
// @Success 200 {object} response.Response "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/users/{id}/roles [put]
func (h *UserApi) UpdateUserRoles(c *gin.Context) {
	svc := userService
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("user service is not available", c)
		return
	}
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req systemRequest.UpdateUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid role request", c)
		return
	}
	err := svc.UpdateUserRoles(c.Request.Context(), actorID(c), id, req.RoleIDs, auditMeta(c))
	if errors.Is(err, systemService.ErrBuiltinAdminProtected) || errors.Is(err, systemService.ErrUserNotFound) || errors.Is(err, systemService.ErrRoleNotFound) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err != nil {
		response.FailWithMessage("failed to update user roles", c)
		return
	}
	response.Ok(c)
}

func parseUintParam(c *gin.Context, name string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		response.FailWithMessage("invalid id", c)
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
		response.FailWithMessage(err.Error(), c)
		return
	}
	if errors.Is(err, systemService.ErrBuiltinAdminProtected) || errors.Is(err, systemService.ErrLDAPProfileReadOnly) {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err != nil {
		response.FailWithMessage("failed to update user", c)
		return
	}
	response.OkWithData(user, c)
}
