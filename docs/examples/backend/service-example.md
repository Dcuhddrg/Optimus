# Service 示例

> 改写自 [gin-vue-admin / backend/service-example.md](https://github.com/flipped-aurora/gin-vue-admin/blob/cee04b08d1262036ca33e3026e1e13d29471c7d5/aiDoc/examples/backend/service-example.md)（`cee04b08`）。2026-10-30 适配 Optimus：保留上游讲解结构，代码、路径及现状说明已修改；上游许可见 [LICENSE.gin-vue-admin](../LICENSE.gin-vue-admin)。

## 这个文件负责什么

Service 处理业务规则、事务边界和数据访问协作，不接收 gin.Context，不组织 HTTP 响应。按现有模块构造函数注入依赖，Store 仍归属于 Service。

## 什么时候应该这样写

- 新增查询、变更或跨操作编排。
- 需要持久化或外部协议适配。
- 补错误、隔离和审计回归。

## 推荐写法示例

以下是审计 Service 的真实方法摘录；`systemModel` 是 `optimus/server/model/system`，`context` 来自标准库：

```go
func NewAuditService(store AuditStore) *AuditService {
    return &AuditService{store: store}
}

func (s *AuditService) QueryLogs(ctx context.Context, f AuditFilter) (systemModel.AuditPage, error) {
    return s.store.QueryLogs(ctx, f)
}
```

此查询本身是委托，不需要为“看起来分层”添加空业务逻辑。动态过滤、Count 和分页在 `service/system/store/audit_store_gorm.go`；用户状态/角色权限变更的业务保护与审计在对应 Service。

## 为什么这样写

返回结果与 error，API 决定 HTTP。构造注入便于用替身测试失败路径；不因上游使用全局 GVA_DB 就改掉已有注入。

## 常见错误

- 把 HTTP 校验、绑定和响应放进 Service。
- 给已提交业务追加日志失败后，谎报业务未发生。
- 仅测试成功路径，忽略最后管理员保护、权限拒绝或快照失败。

## 真实参考文件

- [审计 Service](../../../server/service/system/audit.go)、[Store](../../../server/service/system/store/audit_store_gorm.go)。
- [用户业务](../../../server/service/system/user.go)、[角色业务](../../../server/service/system/role.go)。
- [审计单测](../../../server/service/system/audit_test.go)、[记录质量故障注入](../../../server/service/system/audit_record_quality_test.go)。
