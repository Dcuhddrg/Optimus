package cloud

import "optimus/server/service"

// ApiGroup groups all cloud HTTP APIs.
type ApiGroup struct {
	AccountApi
	AssetApi
}

// Service aliases are refreshed after startup publishes the service group.
var (
	accountService = service.ServiceGroupApp.CloudServiceGroup.CloudAccountService
	assetService   = service.ServiceGroupApp.CloudServiceGroup.AssetRegistryService
)
