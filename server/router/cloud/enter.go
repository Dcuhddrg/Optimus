package cloud

import (
	"github.com/gin-gonic/gin"
	api "optimus/server/api/v1"
)

type RouterGroup struct{}

var (
	accountApi = &api.ApiGroupApp.CloudApiGroup.AccountApi
	assetApi   = &api.ApiGroupApp.CloudApiGroup.AssetApi
)

func (r *RouterGroup) InitCloudRouter(privateGroup *gin.RouterGroup) {

	privateGroup.GET("/cloud/accounts", accountApi.ListAccounts)
	privateGroup.GET("/cloud/accounts/:id", accountApi.GetAccount)
	privateGroup.POST("/cloud/accounts", accountApi.CreateAccount)
	privateGroup.PUT("/cloud/accounts/:id", accountApi.UpdateAccount)
	privateGroup.PUT("/cloud/accounts/:id/tags", accountApi.UpdateAccountTags)
	privateGroup.DELETE("/cloud/accounts/:id", accountApi.DeleteAccount)
	privateGroup.POST("/cloud/accounts/:id/sync", accountApi.SyncAccount)

	privateGroup.GET("/cloud/assets", assetApi.ListAssets)
}
