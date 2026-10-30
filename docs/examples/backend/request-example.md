# Request 示例

## 这个文件负责什么

请求结构负责描述输入，查询条件负责表达筛选与分页，不负责数据库关联或业务编排。Optimus 请求/响应 DTO 当前位于 `server/model/<module>/dto.go`，不是上游独立的 `request/` 包。

## 什么时候应该这样写

- 新增 JSON 创建/更新请求。
- 新增 Query 筛选或分页。
- 明确路径项目、Header 项目选择和服务端授权的关系。

## 推荐写法示例

真实审计请求（HTTP 形态）：

```http
GET /api/v1/system/audit-logs?page=2&pageSize=20&result=failure
Authorization: Bearer <当前登录令牌>
```

下列是 `AuditFilter` 的字段摘录，位于 Service 包；HTTP 层的 `parseAuditFilter` 负责转换，不是把持久化实体直接当请求：

```go
type AuditFilter struct {
    ActorUserID  *uint
    Action       *string
    ResourceType *string
    Result       *string
    StartTime    *time.Time
    EndTime      *time.Time
    Keyword      *string
    Page         int
    PageSize     int
}
```

实际默认 `page=1`、`pageSize=20`，非法条数回退 20，最大 200。时间按 RFC3339 解析；关键词只匹配 detail。登录 JSON 则由 `model/system.LoginRequest` 与 `ShouldBindJSON` 承接。

## 为什么这样写

按真实来源选择绑定方式；查询参数不机械放进 JSON body。指针筛选字段区分“不筛选”和有效值。上游 PageInfo 不在这里假定存在。

## 常见错误

- 在请求 DTO 放数据库关联或未经契约声明的字段。
- 将 `X-Project-Code` 当成授权凭证。
- 用 UI 的校验代替服务端权限检查。

## 真实参考文件

- [请求 DTO](../../../server/model/system/dto.go)、[审计条件](../../../server/service/system/audit.go)。
- [Query 解析](../../../server/api/v1/system/audit.go)、[解析回归](../../../server/api/v1/system/handlers_test.go)。
- [前后端契约](../../frontend-backend/boundary.md)。
