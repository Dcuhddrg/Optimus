package cloud

import "github.com/gin-gonic/gin"

type AccountRouter struct{}

func (r *AccountRouter) InitAccountRouter(privateGroup *gin.RouterGroup) {
	accountRouter := privateGroup.Group("cloud/accounts")
	accountRouter.GET("", accountApi.ListAccounts)
	accountRouter.GET(":id", accountApi.GetAccount)
	accountRouter.POST("", accountApi.CreateAccount)
	accountRouter.PUT(":id", accountApi.UpdateAccount)
	accountRouter.PUT(":id/tags", accountApi.UpdateAccountTags)
	accountRouter.DELETE(":id", accountApi.DeleteAccount)
	accountRouter.POST(":id/sync", accountApi.SyncAccount)
}
