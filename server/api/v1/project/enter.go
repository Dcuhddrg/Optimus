package project

import "optimus/server/service"

type ApiGroup struct {
	ProjectApi
	MemberApi
}

// Each alias points to its service slot so startup can publish the service.
var (
	projectService = &service.ServiceGroupApp.ProjectServiceGroup.ProjectService
)
