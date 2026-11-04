package system

import "github.com/gin-gonic/gin"

type AuditRouter struct{}

func (r *AuditRouter) InitAuditRouter(privateGroup *gin.RouterGroup) {
	auditRouter := privateGroup.Group("system/audit-logs")
	auditRouter.GET("", auditApi.QueryAuditLogs)
	auditRouter.GET("export", auditApi.ExportAuditLogs)
}
