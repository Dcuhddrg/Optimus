package cloud

import "github.com/gin-gonic/gin"

type AccountRouter struct{}

func (r *AccountRouter) InitAccountRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/cloud/accounts", accountApi.ListAccounts)
	privateGroup.GET("/cloud/accounts/:id", accountApi.GetAccount)
	privateGroup.POST("/cloud/accounts", accountApi.CreateAccount)
	privateGroup.PUT("/cloud/accounts/:id", accountApi.UpdateAccount)
	privateGroup.PUT("/cloud/accounts/:id/tags", accountApi.UpdateAccountTags)
	privateGroup.DELETE("/cloud/accounts/:id", accountApi.DeleteAccount)
	privateGroup.POST("/cloud/accounts/:id/sync", accountApi.SyncAccount)
}
