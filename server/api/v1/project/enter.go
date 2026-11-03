package project

import "optimus/server/service"

type ApiGroup struct {
	ProjectApi
	MemberApi
}

// Aliases expose the grouped project service.
var (
	projectService = service.ServiceGroupApp.ProjectServiceGroup.ProjectService
)
