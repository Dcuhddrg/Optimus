package cloud

import "optimus/server/service"

// ApiGroup groups all cloud HTTP APIs.
type ApiGroup struct {
	AccountApi
	AssetApi
}

// Each alias points to its service slot so startup can publish the service.
var (
	accountService = &service.ServiceGroupApp.CloudServiceGroup.CloudAccountService
	assetService = &service.ServiceGroupApp.CloudServiceGroup.AssetRegistryService
)
