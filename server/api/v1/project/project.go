package project

import (
	"optimus/server/global"
	"optimus/server/model/common/response"
	projectRequest "optimus/server/model/project/request"
	projectResponse "optimus/server/model/project/response"

	"github.com/gin-gonic/gin"
)

type ProjectApi struct{}

// swag 只能解析本文件导入的包别名；保留此引用，供 Swagger 注释使用 projectResponse。
var _ = projectResponse.ProjectView{}

// @Tags Project
// @Summary 当前用户可访问的项目列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=object{list=[]projectResponse.ProjectView}} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects [get]
func (h *ProjectApi) ListProjects(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	views, err := svc.List(c.Request.Context(), projectActorID(c))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.OkWithData(gin.H{"list": views}, c)
}

// @Tags Project
// @Summary 创建项目
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body projectRequest.CreateProjectRequest true "请求体"
// @Success 200 {object} response.Response{data=projectResponse.ProjectView} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects [post]
func (h *ProjectApi) CreateProject(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectRequest.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid project payload", c)
		return
	}
	view, err := svc.Create(c.Request.Context(), projectActorID(c), req)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.OkWithData(view, c)
}

// @Tags Project
// @Summary 获取项目详情
// @Security ApiKeyAuth
// @Produce application/json
// @Param code path string true "项目编码"
// @Success 200 {object} response.Response{data=projectResponse.ProjectView} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects/{code} [get]
func (h *ProjectApi) GetProject(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	view, err := svc.Get(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.OkWithData(view, c)
}

// @Tags Project
// @Summary 更新项目
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param code path string true "项目编码"
// @Param data body projectRequest.UpdateProjectRequest true "请求体"
// @Success 200 {object} response.Response{data=projectResponse.ProjectView} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects/{code} [put]
func (h *ProjectApi) UpdateProject(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectRequest.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid project payload", c)
		return
	}
	view, err := svc.Update(c.Request.Context(), projectActorID(c), c.Param("code"), req)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	response.OkWithData(view, c)
}

// @Tags Project
// @Summary 删除项目
// @Security ApiKeyAuth
// @Produce application/json
// @Param code path string true "项目编码"
// @Success 200 {object} response.Response "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /projects/{code} [delete]
func (h *ProjectApi) DeleteProject(c *gin.Context) {
	svc := projectService
	if global.OPTIMUS_DB == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	if err := svc.Delete(c.Request.Context(), projectActorID(c), c.Param("code")); err != nil {
		writeProjectError(c, err)
		return
	}
	response.Ok(c)
}
