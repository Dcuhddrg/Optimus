package cloud

import (
	cloudApi "optimus/server/api/v1/cloud"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct{}

func (r *RouterGroup) InitCloudRouter(privateGroup *gin.RouterGroup, handlers *cloudApi.CloudHandlers) {
	if handlers == nil {
		handlers = cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{})
	}

	privateGroup.GET("/cloud/accounts", handlers.Account.ListAccounts)
	privateGroup.GET("/cloud/accounts/:id", handlers.Account.GetAccount)
	privateGroup.POST("/cloud/accounts", handlers.Account.CreateAccount)
	privateGroup.PUT("/cloud/accounts/:id", handlers.Account.UpdateAccount)
	privateGroup.PUT("/cloud/accounts/:id/tags", handlers.Account.UpdateAccountTags)
	privateGroup.DELETE("/cloud/accounts/:id", handlers.Account.DeleteAccount)
	privateGroup.POST("/cloud/accounts/:id/sync", handlers.Account.SyncAccount)

	privateGroup.GET("/cloud/assets", handlers.Asset.ListAssets)
}
