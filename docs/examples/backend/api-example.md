# API 示例

## 这个文件负责什么

API 接收 HTTP 输入、校验来源、调用 Service 并组织响应；Swagger 需准确描述实际路由和格式。

## 什么时候应该这样写

- 新增接口或调整参数。
- 对接客户端类型与错误处理。
- 为下载等非 JSON 场景声明例外。

## 推荐写法示例

审计 handler 真实摘录（所在文件已有 Gin/http、systemModel 及 response 导入）：

```go
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
```

Query 用 `c.Query`/`DefaultQuery`，JSON body 才用 `ShouldBindJSON`，路径用 `c.Param`；Header/表单/文件按实际协议选择。审计 handler 使用本仓库新增的 `utils/response`，成功为 `{code:0,data:{list,total,page,pageSize},msg}`，失败保留 HTTP 状态并输出 envelope。CSV 成功仍为文本；其他未登记端点继续使用旧格式。完整 Swagger 注释见源文件，摘录省略注释不代表所有端点均已核查。

## 为什么这样写

参数来源与协议一致；领域错误映射到已有 HTTP 语义，不把所有失败改成 HTTP 200。统一 envelope 的目标见现有契约，不在示例任务中实施。

## 常见错误

- 复制 GVA 的 response 包导入，而本仓库并无该封装。
- 机械地给查询或导出绑定 JSON body。
- CSV 包 JSON、泄漏原始 SQL 或把注释当实际验收。

## 真实参考文件

- [审计 API](../../../server/api/v1/system/audit.go)、[handler 构造](../../../server/api/v1/system/handlers.go)、[公共响应工具](../../../server/utils/response/response.go)。
- [handler 回归](../../../server/api/v1/system/handlers_test.go)、[接口契约](../../frontend-backend/boundary.md)。
