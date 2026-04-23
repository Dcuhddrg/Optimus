package system

import (
	systemApi "optimus/server/api/v1/system"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct{}

func (r *RouterGroup) InitSystemRouter(publicGroup *gin.RouterGroup, _ *gin.RouterGroup) {
	api := &systemApi.Api{}

	publicGroup.GET("/health", api.Health)
	publicGroup.GET("/ping", api.Ping)
	publicGroup.GET("/base/ping", api.Ping)
}
