package system

import (
	"net/http"

	"optimus/server/global"
	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

func (a *Api) AuditLogs(c *gin.Context) {
	if global.OPTIMUS_DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not available"})
		return
	}
	service := systemService.NewAuditService(systemService.NewGormAuditStore(global.OPTIMUS_DB))
	result, err := service.ListAuditLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list audit logs"})
		return
	}
	c.JSON(http.StatusOK, result)
}
