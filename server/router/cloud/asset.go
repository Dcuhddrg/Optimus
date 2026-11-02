package cloud

import "github.com/gin-gonic/gin"

type AssetRouter struct{}

func (r *AssetRouter) InitAssetRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/cloud/assets", assetApi.ListAssets)
}
