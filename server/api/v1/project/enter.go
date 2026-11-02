package project

import "optimus/server/service"

type ApiGroup struct {
	ProjectApi
	MemberApi
}

var projectServices = &service.ServiceGroupApp.ProjectServiceGroup
