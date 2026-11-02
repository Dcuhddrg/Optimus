package project

import "github.com/gin-gonic/gin"

type MemberRouter struct{}

func (r *MemberRouter) InitMemberRouter(rg *gin.RouterGroup) {
	group := rg.Group("projects")
	group.GET(":code/members", memberApi.ListMembers)
	group.POST(":code/members", memberApi.AddMember)
	group.PATCH(":code/members/:userId", memberApi.UpdateMemberRole)
	group.DELETE(":code/members/:userId", memberApi.RemoveMember)
}
