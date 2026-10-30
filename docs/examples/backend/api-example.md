# API 示例

> 改写自 [gin-vue-admin / backend/api-example.md](https://github.com/flipped-aurora/gin-vue-admin/blob/cee04b08d1262036ca33e3026e1e13d29471c7d5/aiDoc/examples/backend/api-example.md)（`cee04b08`）。2026-10-30 适配 Optimus：保留上游讲解结构，代码、路径及现状说明已修改；上游许可见 [LICENSE.gin-vue-admin](../LICENSE.gin-vue-admin)。

## 这个文件负责什么

API 接收 HTTP 输入、校验来源、调用 Service 并组织响应；Swagger 需准确描述实际路由和格式。

## 什么时候应该这样写

- 新增接口或调整参数。
- 对接客户端类型与错误处理。
- 为下载等非 JSON 场景声明例外。

## 推荐写法示例

审计 handler 真实摘录（所在文件已有 Gin/http 导入）：

```go
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
```

Query 用 `c.Query`/`DefaultQuery`，JSON body 才用 `ShouldBindJSON`，路径用 `c.Param`；Header/表单/文件按实际协议选择。当前 handler 的 c.JSON 是存量实现，不是上游 response 包示例。当前成功 JSON 是原始审计分页 DTO，错误 `{message}`；Swagger 完整性仍应逐端点核对，不能把上面没有注释的摘录当成已完成 Swagger 的证明。

## 为什么这样写

参数来源与协议一致；领域错误映射到已有 HTTP 语义，不把所有失败改成 HTTP 200。统一 envelope 的目标见现有契约，不在示例任务中实施。

## 常见错误

- 复制 GVA 的 response 包导入，而本仓库并无该封装。
- 机械地给查询或导出绑定 JSON body。
- CSV 包 JSON、泄漏原始 SQL 或把注释当实际验收。

## 真实参考文件

- [审计 API](../../../server/api/v1/system/audit.go)、[handler 构造](../../../server/api/v1/system/handlers.go)。
- [handler 回归](../../../server/api/v1/system/handlers_test.go)、[接口契约](../../frontend-backend/boundary.md)。
