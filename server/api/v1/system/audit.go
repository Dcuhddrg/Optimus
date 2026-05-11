package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *AuditHandler) AuditLogs(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "audit service is not available")
		return
	}
	result, err := h.service.ListAuditLogs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list audit logs"})
		return
	}
	c.JSON(http.StatusOK, result)
}
