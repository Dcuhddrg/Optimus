package system

import "github.com/gin-gonic/gin"

type BaseRouter struct{}

func (r *BaseRouter) InitBaseRouter(publicGroup *gin.RouterGroup) {
	publicGroup.GET("/health", healthApi.Health)
	publicGroup.GET("/ping", baseApi.Ping)
	publicGroup.GET("/base/ping", baseApi.Ping)
}
