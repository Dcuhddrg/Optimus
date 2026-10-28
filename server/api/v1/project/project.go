package project

import (
	"net/http"

	projectModel "optimus/server/model/project"

	"github.com/gin-gonic/gin"
)

func (h *ProjectHandlers) ListProjects(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	views, err := h.svc.List(c.Request.Context(), projectActorID(c))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

func (h *ProjectHandlers) CreateProject(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectModel.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project payload"})
		return
	}
	view, err := h.svc.Create(c.Request.Context(), projectActorID(c), req)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *ProjectHandlers) GetProject(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	view, err := h.svc.Get(c.Request.Context(), projectActorID(c), c.Param("code"))
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *ProjectHandlers) UpdateProject(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	var req projectModel.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid project payload"})
		return
	}
	view, err := h.svc.Update(c.Request.Context(), projectActorID(c), c.Param("code"), req)
	if err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *ProjectHandlers) DeleteProject(c *gin.Context) {
	if h.svc == nil {
		writeProjectServiceUnavailable(c)
		return
	}
	if err := h.svc.Delete(c.Request.Context(), projectActorID(c), c.Param("code")); err != nil {
		writeProjectError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}
