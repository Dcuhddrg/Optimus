package project

import (
	"github.com/gin-gonic/gin"
	api "optimus/server/api/v1"
)

type RouterGroup struct{}

var (
	projectApi = &api.ApiGroupApp.ProjectApiGroup.ProjectApi
	memberApi  = &api.ApiGroupApp.ProjectApiGroup.MemberApi
)

func (r *RouterGroup) InitProjectRouter(rg *gin.RouterGroup) {
	group := rg.Group("projects")
	{
		group.GET("", projectApi.ListProjects)
		group.POST("", projectApi.CreateProject)
		group.GET(":code", projectApi.GetProject)
		group.PUT(":code", projectApi.UpdateProject)
		group.DELETE(":code", projectApi.DeleteProject)

		group.GET(":code/members", memberApi.ListMembers)
		group.POST(":code/members", memberApi.AddMember)
		group.PATCH(":code/members/:userId", memberApi.UpdateMemberRole)
		group.DELETE(":code/members/:userId", memberApi.RemoveMember)
	}
}
