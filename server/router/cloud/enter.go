package cloud

import (
	api "optimus/server/api/v1"
)

type RouterGroup struct {
	AccountRouter
	AssetRouter
}

var (
	accountApi = &api.ApiGroupApp.CloudApiGroup.AccountApi
	assetApi   = &api.ApiGroupApp.CloudApiGroup.AssetApi
)
