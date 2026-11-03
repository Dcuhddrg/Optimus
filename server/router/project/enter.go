package project

import (
	api "optimus/server/api/v1"
)

type RouterGroup struct {
	ProjectRouter
	MemberRouter
}

var (
	projectApi = api.ApiGroupApp.ProjectApiGroup.ProjectApi
	memberApi  = api.ApiGroupApp.ProjectApiGroup.MemberApi
)
