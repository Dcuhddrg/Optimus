package service

import (
	"optimus/server/service/cloud"
	"optimus/server/service/project"
	"optimus/server/service/system"
)

// ServiceGroupApp keeps stable module addresses for API aliases and workers.
// Dependencies are assembled once by initialize before routes are registered.
var ServiceGroupApp = &ServiceGroup{
	SystemServiceGroup: system.ServiceGroup{RBAC: (*system.RBACService)(nil)},
}

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ProjectServiceGroup project.ServiceGroup
	CloudServiceGroup   cloud.ServiceGroup
}
