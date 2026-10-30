package system

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"
	"optimus/server/utils/response"

	"github.com/gin-gonic/gin"
)

// QueryAuditLogs returns the migrated audit pagination contract.
// @Tags System Audit
// @Summary Query audit logs
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param page query int false "Page, starting from 1" default(1)
// @Param pageSize query int false "Page size, 1 to 200" default(20)
// @Param actorUserId query int false "Actor user ID"
// @Param action query string false "Exact action code"
// @Param resourceType query string false "Exact resource type"
// @Param result query string false "Result"
// @Param startTime query string false "Start time in RFC3339"
// @Param endTime query string false "End time in RFC3339"
// @Param keyword query string false "Recorded detail keyword"
// @Success 200 {object} response.Envelope{data=systemModel.AuditListData}
// @Failure 401 {object} response.Envelope
// @Failure 403 {object} response.Envelope
// @Failure 500 {object} response.Envelope
// @Failure 503 {object} response.Envelope
// @Router /system/audit-logs [get]
func (h *AuditHandler) QueryAuditLogs(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "audit service is not available")
		return
	}
	result, err := h.service.QueryLogs(c.Request.Context(), parseAuditFilter(c))
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to query audit logs")
		return
	}
	list := result.Items
	if list == nil {
		list = make([]systemModel.AuditLogEntry, 0)
	}
	response.Success(c, http.StatusOK, systemModel.AuditListData{List: list, Total: result.Total, Page: result.Page, PageSize: result.PageSize})
}

// ExportAuditLogs retains CSV success and uses migrated JSON failures.
// @Tags System Audit
// @Summary Export matching audit logs, up to 10000 rows
// @Produce text/csv
// @Param Authorization header string true "Bearer token"
// @Param actorUserId query int false "Actor user ID"
// @Param action query string false "Exact action code"
// @Param resourceType query string false "Exact resource type"
// @Param result query string false "Result"
// @Param startTime query string false "Start time in RFC3339"
// @Param endTime query string false "End time in RFC3339"
// @Param keyword query string false "Recorded detail keyword"
// @Success 200 {string} string "CSV attachment"
// @Failure 401 {object} response.Envelope
// @Failure 403 {object} response.Envelope
// @Failure 500 {object} response.Envelope
// @Failure 503 {object} response.Envelope
// @Router /system/audit-logs/export [get]
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	if h.service == nil {
		response.Failure(c, http.StatusServiceUnavailable, "audit service is not available")
		return
	}
	entries, err := h.service.ExportLogs(c.Request.Context(), parseAuditFilter(c))
	if err != nil {
		response.Failure(c, http.StatusInternalServerError, "failed to export audit logs")
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="audit-logs.csv"`)

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	if err := w.Write([]string{"id", "actor_username", "action", "resource_type", "resource_id", "result", "ip", "created_at", "detail"}); err != nil {
		return
	}
	for _, entry := range entries {
		if err := w.Write([]string{
			strconv.FormatUint(uint64(entry.ID), 10),
			entry.ActorUsername,
			entry.Action,
			entry.ResourceType,
			entry.ResourceID,
			entry.Result,
			entry.IP,
			entry.CreatedAt.Format(time.RFC3339),
			entry.Detail,
		}); err != nil {
			return
		}
	}
}

func parseAuditFilter(c *gin.Context) systemService.AuditFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	f := systemService.AuditFilter{Page: page, PageSize: pageSize}

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
