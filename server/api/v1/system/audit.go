package system

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	systemUsecase "optimus/server/service/system/usecase"

	"github.com/gin-gonic/gin"
)

func (h *AuditHandler) QueryAuditLogs(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "audit service is not available")
		return
	}
	result, err := h.service.QueryLogs(c.Request.Context(), parseAuditFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query audit logs"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "audit service is not available")
		return
	}
	entries, err := h.service.ExportLogs(c.Request.Context(), parseAuditFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to export audit logs"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="audit-logs.csv"`)

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"id", "actor_username", "action", "resource_type", "resource_id", "result", "ip", "created_at", "detail"})
	for _, entry := range entries {
		_ = w.Write([]string{
			strconv.FormatUint(uint64(entry.ID), 10),
			entry.ActorUsername,
			entry.Action,
			entry.ResourceType,
			entry.ResourceID,
			entry.Result,
			entry.IP,
			entry.CreatedAt.Format(time.RFC3339),
			entry.Detail,
		})
	}
	w.Flush()
}

func parseAuditFilter(c *gin.Context) systemUsecase.AuditFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	f := systemUsecase.AuditFilter{Page: page, PageSize: pageSize}

	if v := c.Query("actorUserId"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			uid := uint(id)
			f.ActorUserID = &uid
		}
	}
	if v := c.Query("action"); v != "" {
		f.Action = &v
	}
	if v := c.Query("resourceType"); v != "" {
		f.ResourceType = &v
	}
	if v := c.Query("result"); v != "" {
		f.Result = &v
	}
	if v := c.Query("startTime"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.StartTime = &t
		}
	}
	if v := c.Query("endTime"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.EndTime = &t
		}
	}
	if v := c.Query("keyword"); v != "" {
		f.Keyword = &v
	}
	return f
}
