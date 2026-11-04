package cloud

import "optimus/server/service"

// ApiGroup groups all cloud HTTP APIs.
type ApiGroup struct {
	AccountApi
	AssetApi
}

// Service aliases copy stateless service values, as in GVA.
var (
	accountService = service.ServiceGroupApp.CloudServiceGroup.CloudAccountService
	assetService   = service.ServiceGroupApp.CloudServiceGroup.AssetRegistryService
)
