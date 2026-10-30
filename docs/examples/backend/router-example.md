# Router 示例

## 这个文件负责什么

Router 负责分组、中间件继承和 handler 绑定，不访问数据库或编排业务。

## 什么时候应该这样写

- 新增模块端点。
- 区分公开与私有入口。
- 为需要项目上下文的资源路由挂 ProjectScope。

## 推荐写法示例

`InitSystemRouter` 内的真实绑定摘录：

```go
privateGroup.GET("/system/audit-logs", handlers.Audit.QueryAuditLogs)
privateGroup.GET("/system/audit-logs/export", handlers.Audit.ExportAuditLogs)
```

外层 `/api/v1` 及 JWT/RBAC 在 initialize 装配；系统恢复入口不受无关项目选择阻断，ProjectScope 只装到需要资源项目范围的云组。不要复制上游 `middleware.OperationRecord()`：本仓库的关键审计发生在 Service，相关注册点需按当前实现核对。

## 为什么这样写

读取和导出都沿用服务端权限边界，导出 API 权限不等于菜单可见。登录属于公开请求，恢复/私有 401 的客户端语义见 #27。

## 常见错误

- 在路由里做业务。
- 为方便新增平行注册链或放宽私有端点。
- 把 ProjectScope 挂到所有私有组，或假定存在上游 OperationRecord。

## 真实参考文件

- [系统路由](../../../server/router/system/enter.go)、[初始化](../../../server/initialize/router.go)。
- [作用域回归](../../../server/initialize/router_test.go)、[中间件](../../../server/middleware/project_scope.go)。
