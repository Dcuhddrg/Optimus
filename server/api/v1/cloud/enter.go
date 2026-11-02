package cloud

import "optimus/server/service"

// AccountApi handles HTTP requests for cloud accounts.
type AccountApi struct{}

// AssetApi handles HTTP requests for cloud assets.
type AssetApi struct{}

// ApiGroup groups all cloud HTTP APIs.
type ApiGroup struct {
	AccountApi
	AssetApi
}

var cloudServices = &service.ServiceGroupApp.CloudServiceGroup
