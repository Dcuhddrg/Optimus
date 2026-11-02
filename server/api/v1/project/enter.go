package project

import "optimus/server/service"

type ProjectApi struct{}
type MemberApi struct{}
type ApiGroup struct {
	ProjectApi
	MemberApi
}

var projectServices = &service.ServiceGroupApp.ProjectServiceGroup
