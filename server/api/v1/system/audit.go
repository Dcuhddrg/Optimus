package system

import (
	"encoding/csv"
	"optimus/server/global"
	commonRequest "optimus/server/model/common/request"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"
	"strconv"
	"time"

	"optimus/server/model/common/response"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

type AuditApi struct{}

// QueryAuditLogs returns the migrated audit pagination contract.
// @Tags Audit
// @Summary 分页查询审计日志
// @Security ApiKeyAuth
// @Produce application/json
// @Param page query int false "Page, starting from 1" default(1)
// @Param pageSize query int false "Page size, 1 to 200" default(20)
// @Param actorUserId query int false "Actor user ID"
// @Param action query string false "Exact action code"
// @Param resourceType query string false "Exact resource type"
// @Param result query string false "Result"
// @Param startTime query string false "Start time in RFC3339"
// @Param endTime query string false "End time in RFC3339"
// @Param keyword query string false "Recorded detail keyword"
// @Success 200 {object} response.Response{data=systemResponse.AuditListData} "code 0 成功，code 7 业务失败"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/audit-logs [get]
func (h *AuditApi) QueryAuditLogs(c *gin.Context) {
	svc := auditService
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("audit service is not available", c)
		return
	}
	result, err := svc.QueryLogs(c.Request.Context(), parseAuditFilter(c))
	if err != nil {
		response.FailWithMessage("failed to query audit logs", c)
		return
	}
	list := result.Items
	if list == nil {
		list = make([]systemModel.AuditLogEntry, 0)
	}
	response.OkWithData(systemResponse.AuditListData{List: list, Total: result.Total, Page: result.Page, PageSize: result.PageSize}, c)
}

// ExportAuditLogs retains CSV success and uses migrated JSON failures.
// @Tags Audit
// @Summary 导出符合筛选的审计日志 CSV，最多 10000 条；失败时返回 JSON
// @Security ApiKeyAuth
// @Produce text/csv
// @Param actorUserId query int false "Actor user ID"
// @Param action query string false "Exact action code"
// @Param resourceType query string false "Exact resource type"
// @Param result query string false "Result"
// @Param startTime query string false "Start time in RFC3339"
// @Param endTime query string false "End time in RFC3339"
// @Param keyword query string false "Recorded detail keyword"
// @Success 200 {string} string "CSV 附件；失败时为 JSON {code:7}"
// @Failure 401 {object} response.Response "未认证"
// @Router /system/audit-logs/export [get]
func (h *AuditApi) ExportAuditLogs(c *gin.Context) {
	svc := auditService
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("audit service is not available", c)
		return
	}
	entries, err := svc.ExportLogs(c.Request.Context(), parseAuditFilter(c))
	if err != nil {
		response.FailWithMessage("failed to export audit logs", c)
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
	pagination := commonRequest.PageInfo{Page: page, PageSize: pageSize}
	pagination.Normalize()

	f := systemRequest.AuditSearch{PageInfo: pagination}

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
	return systemService.AuditFilter{
		ActorUserID: f.ActorUserID, Action: f.Action, ResourceType: f.ResourceType,
		Result: f.Result, StartTime: f.StartTime, EndTime: f.EndTime, Keyword: f.Keyword,
		Page: f.Page, PageSize: f.PageSize,
	}
}
