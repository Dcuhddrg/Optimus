package cloud

import "github.com/gin-gonic/gin"

type AssetRouter struct{}

func (r *AssetRouter) InitAssetRouter(privateGroup *gin.RouterGroup) {
	assetRouter := privateGroup.Group("cloud/assets")
	assetRouter.GET("", assetApi.ListAssets)
}
