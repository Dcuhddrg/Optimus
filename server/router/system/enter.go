package system

import (
	systemApi "optimus/server/api/v1/system"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct{}

func (r *RouterGroup) InitSystemRouter(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
	api := &systemApi.Api{}

	publicGroup.GET("/health", api.Health)
	publicGroup.GET("/ping", api.Ping)
	publicGroup.GET("/base/ping", api.Ping)
	publicGroup.POST("/system/auth/login", api.Login)

	privateGroup.GET("/system/auth/me", api.CurrentUser)
	privateGroup.POST("/system/auth/logout", api.Logout)
	privateGroup.GET("/system/audit-logs", api.AuditLogs)
}
