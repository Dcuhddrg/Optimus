package cloud

import "optimus/server/service"

// RefreshServiceAliases publishes startup dependencies to the value aliases.
// Call only after service groups are initialized and before requests are served.
func RefreshServiceAliases() {
	accountService = service.ServiceGroupApp.CloudServiceGroup.CloudAccountService
	assetService = service.ServiceGroupApp.CloudServiceGroup.AssetRegistryService
}
