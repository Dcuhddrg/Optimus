package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *Api) AuditLogs(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []any{}, "total": 0})
}
