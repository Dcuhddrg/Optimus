package project

import "github.com/gin-gonic/gin"

type ProjectRouter struct{}

func (r *ProjectRouter) InitProjectRouter(rg *gin.RouterGroup) {
	group := rg.Group("projects")
	group.GET("", projectApi.ListProjects)
	group.POST("", projectApi.CreateProject)
	group.GET(":code", projectApi.GetProject)
	group.PUT(":code", projectApi.UpdateProject)
	group.DELETE(":code", projectApi.DeleteProject)
}
