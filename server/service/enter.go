package service

import (
	"optimus/server/service/cloud"
	"optimus/server/service/project"
	"optimus/server/service/system"
)

// ServiceGroupApp aggregates stateless module services, as in GVA. Services read
// shared dependencies (DB, config, clients) from global at call time.
var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ProjectServiceGroup project.ServiceGroup
	CloudServiceGroup   cloud.ServiceGroup
}
