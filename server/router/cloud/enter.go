package cloud

import (
	"github.com/gin-gonic/gin"
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

func (r *RouterGroup) InitCloudRouter(privateGroup *gin.RouterGroup) {
	r.AccountRouter.InitAccountRouter(privateGroup)
	r.AssetRouter.InitAssetRouter(privateGroup)
}
