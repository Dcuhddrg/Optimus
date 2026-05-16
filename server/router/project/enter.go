package project

import (
	projectApi "optimus/server/api/v1/project"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct{}

func (r *RouterGroup) InitProjectRouter(rg *gin.RouterGroup, h *projectApi.ProjectHandlers) {
	group := rg.Group("projects")
	{
		group.GET("", h.ListProjects)
		group.POST("", h.CreateProject)
		group.GET(":code", h.GetProject)
		group.PUT(":code", h.UpdateProject)
		group.DELETE(":code", h.DeleteProject)

		group.GET(":code/members", h.ListMembers)
		group.POST(":code/members", h.AddMember)
		group.PATCH(":code/members/:userId", h.UpdateMemberRole)
		group.DELETE(":code/members/:userId", h.RemoveMember)
	}
}
