package v1

import (
	"optimus/server/api/v1/cloud"
	"optimus/server/api/v1/project"
	"optimus/server/api/v1/system"
)

// ApiGroupApp contains stateless responsibility APIs at stable addresses.
var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	SystemApiGroup  system.ApiGroup
	ProjectApiGroup project.ApiGroup
	CloudApiGroup   cloud.ApiGroup
}
