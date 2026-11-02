package project

import (
	"github.com/gin-gonic/gin"
	api "optimus/server/api/v1"
)

type RouterGroup struct {
	ProjectRouter
	MemberRouter
}

var (
	projectApi = &api.ApiGroupApp.ProjectApiGroup.ProjectApi
	memberApi  = &api.ApiGroupApp.ProjectApiGroup.MemberApi
)

func (r *RouterGroup) InitProjectRouter(rg *gin.RouterGroup) {
	r.ProjectRouter.InitProjectRouter(rg)
	r.MemberRouter.InitMemberRouter(rg)
}
