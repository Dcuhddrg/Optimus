# Model 示例

## 这个文件负责什么

Model 描述实体、请求/响应结构和持久化映射；不承担 Service 的业务编排。字段必须与客户端类型及真实存储一致。

## 什么时候应该这样写

- 新增实体或 DTO。
- 补字段、索引与 JSON 标签。
- 需要区分数据库实体和可见响应。

## 推荐写法示例

来自 `server/model/system/user.go` 的实体摘录；`time` 为 Go 标准库：

```go
type SystemAuditLog struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    ActorUserID   uint      `gorm:"index" json:"actorUserId"`
    ActorUsername string    `gorm:"size:128" json:"actorUsername"`
    Action        string    `gorm:"size:128;not null;index" json:"action"`
    ResourceType  string    `gorm:"size:128;index" json:"resourceType"`
    ResourceID    string    `gorm:"size:128;index" json:"resourceId"`
    Result        string    `gorm:"size:32;not null;index" json:"result"`
    IP            string    `gorm:"size:64" json:"ip"`
    UserAgent     string    `gorm:"size:255" json:"userAgent"`
    Detail        string    `gorm:"type:text" json:"detail"`
    CreatedAt     time.Time `json:"createdAt"`
}
```

`AuditPage`/`AuditLogEntry` 单独定义在 `dto.go`，供 Store/Service 使用；HTTP 的 `AuditListData` 将 items 映射为 list；不导入上游 `global.GVA_MODEL`。

## 为什么这样写

主键、JSON 名称和时间字段沿用实际模块约定。操作人名称是事件发生时的快照，历史日志不根据当前用户信息回填。

## 常见错误

- 引入不存在的 GVA_MODEL 或把旧结构当当前约束。
- 在 Model 调用 API/Service。
- 把密码、token 或整份请求作为审计 Detail。

## 真实参考文件

- [实体](../../../server/model/system/user.go)、[响应 DTO](../../../server/model/system/dto.go)。
- [实体到响应映射](../../../server/service/system/store/audit_store_gorm.go)、[快照测试](../../../server/service/system/store/audit_store_gorm_test.go)。
