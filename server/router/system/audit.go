package system

import "github.com/gin-gonic/gin"

type AuditRouter struct{}

func (r *AuditRouter) InitAuditRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/system/audit-logs", auditApi.QueryAuditLogs)
	privateGroup.GET("/system/audit-logs/export", auditApi.ExportAuditLogs)
}
