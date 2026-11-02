package cloud

import "optimus/server/service"

// ApiGroup groups all cloud HTTP APIs.
type ApiGroup struct {
	AccountApi
	AssetApi
}

var cloudServices = &service.ServiceGroupApp.CloudServiceGroup
