package project

import (
	"optimus/server/global"
	"optimus/server/model/common/response"
	projectRequest "optimus/server/model/project/request"

	"github.com/gin-gonic/gin"
)

type ProjectApi struct{}

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
