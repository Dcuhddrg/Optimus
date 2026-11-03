package project

import (
	"net/http"
	projectRequest "optimus/server/model/project/request"

	"github.com/gin-gonic/gin"
)

type ProjectApi struct{}

func (h *ProjectApi) ListProjects(c *gin.Context) {
	svc := *projectService
	if svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	views, err := svc.List(c.Request.Context(), projectActorID(c))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

func (h *ProjectApi) CreateProject(c *gin.Context) {
	svc := *projectService
	if svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectRequest.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project payload"})
		return
	}
	view, err := svc.Create(c.Request.Context(), projectActorID(c), req)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *ProjectApi) GetProject(c *gin.Context) {
	svc := *projectService
	if svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	view, err := svc.Get(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *ProjectApi) UpdateProject(c *gin.Context) {
	svc := *projectService
	if svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectRequest.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project payload"})
		return
	}
	view, err := svc.Update(c.Request.Context(), projectActorID(c), c.Param("code"), req)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *ProjectApi) DeleteProject(c *gin.Context) {
	svc := *projectService
	if svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	if err := svc.Delete(c.Request.Context(), projectActorID(c), c.Param("code")); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}
