package project

import (
	"optimus/server/global"
	"optimus/server/model/common/response"
	projectRequest "optimus/server/model/project/request"
	projectResponse "optimus/server/model/project/response"

	"github.com/gin-gonic/gin"
)

type MemberApi struct{}

// swag 只能解析本文件导入的包别名；保留此引用，供 Swagger 注释使用 projectResponse。
var _ = projectResponse.ProjectMember{}

// @Tags ProjectMember
// @Summary 项目成员列表
// @Security ApiKeyAuth
// @Produce application/json
// @Param code path string true "项目编码"
// @Success 200 {object} response.Response{data=object{list=[]projectResponse.ProjectMember}} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects/{code}/members [get]
func (h *MemberApi) ListMembers(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	members, err := svc.ListMembers(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.OkWithData(gin.H{"list": members}, c)
}

// @Tags ProjectMember
// @Summary 添加项目成员
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param code path string true "项目编码"
// @Param data body projectRequest.AddMemberRequest true "请求体"
// @Success 200 {object} response.Response "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects/{code}/members [post]
func (h *MemberApi) AddMember(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectRequest.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid member payload", c)
		return
	}
	if req.UserID == 0 {
		response.FailWithMessage("userId is required", c)
		return
	}
	if err := svc.AddMember(c.Request.Context(), projectActorID(c), c.Param("code"), req.UserID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Ok(c)
}

// @Tags ProjectMember
// @Summary 修改成员的项目角色
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param code path string true "项目编码"
// @Param userId path int true "用户 ID"
// @Param data body projectRequest.UpdateMemberRoleRequest true "请求体"
// @Success 200 {object} response.Response "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects/{code}/members/{userId} [patch]
func (h *MemberApi) UpdateMemberRole(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	var req projectRequest.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid member payload", c)
		return
	}
	if err := svc.UpdateMemberRole(c.Request.Context(), projectActorID(c), c.Param("code"), userID, req.ProjectRole); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Ok(c)
}

// @Tags ProjectMember
// @Summary 移除项目成员
// @Security ApiKeyAuth
// @Produce application/json
// @Param code path string true "项目编码"
// @Param userId path int true "用户 ID"
// @Success 200 {object} response.Response "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects/{code}/members/{userId} [delete]
func (h *MemberApi) RemoveMember(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	userID, ok := projectParseUintParam(c, "userId")
	if !ok {
		return
	}
	if err := svc.RemoveMember(c.Request.Context(), projectActorID(c), c.Param("code"), userID); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Ok(c)
}
