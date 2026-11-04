package system

import "github.com/gin-gonic/gin"

type BaseRouter struct{}

func (r *BaseRouter) InitBaseRouter(publicGroup *gin.RouterGroup) {
	publicGroup.GET("/health", healthApi.Health)
	publicGroup.GET("/ping", baseApi.Ping)
	baseRouter := publicGroup.Group("base")
	baseRouter.GET("ping", baseApi.Ping)
}
