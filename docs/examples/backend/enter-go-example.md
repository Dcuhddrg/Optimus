# enter.go 示例

## 这个文件负责什么

`enter.go` 只做聚合：模块级把职责类型匿名嵌入 Group 并声明包级别名，根级把模块 Group 汇总为 `ServiceGroupApp`、`ApiGroupApp`、`RouterGroupApp`。职责类型本身写在各自文件里。

## 什么时候应该这样写

- 新增模块
- 模块内新增一个职责（Service、Api 或 Router）

## 推荐写法示例

### 根级聚合

```go
// server/service/enter.go
package service

import (
	"optimus/server/service/cloud"
	"optimus/server/service/order"
	"optimus/server/service/project"
	"optimus/server/service/system"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	SystemServiceGroup  system.ServiceGroup
	ProjectServiceGroup project.ServiceGroup
	CloudServiceGroup   cloud.ServiceGroup
	OrderServiceGroup   order.ServiceGroup
}
```

### 模块级聚合

```go
// server/service/order/enter.go
package order

type ServiceGroup struct {
	OrderService
}
```

```go
// server/api/v1/order/enter.go
package order

import "optimus/server/service"

type ApiGroup struct {
	OrderApi
}

var orderSvc = service.ServiceGroupApp.OrderServiceGroup.OrderService
```

```go
// server/router/order/enter.go
package order

import api "optimus/server/api/v1"

type RouterGroup struct {
	OrderRouter
}

var orderApi = api.ApiGroupApp.OrderApiGroup.OrderApi
```

## 为什么这样写

- 与 GVA 相同：Group 匿名嵌入职责类型，别名在包初始化时复制值；Service 无状态，所以复制的值始终可用，不需要启动后刷新。
- 上层只依赖别名，不在 handler 里拼接 `ServiceGroupApp...` 长路径。
- `initialize` 只调用 `InitXxxRouter`，不另建装配或注册链。

## 常见错误

- 在 `enter.go` 里写职责类型或业务逻辑（契约测试 `TestEnterFilesOnlyDeclareGroupTypes` 会失败）
- 新模块只加了模块级 Group，忘记加入根级 Group
- 在 handler 里直接引用 `service.ServiceGroupApp...`

## 真实参考文件

- [server/service/enter.go](../../../server/service/enter.go)
- [server/api/v1/project/enter.go](../../../server/api/v1/project/enter.go)
- [server/router/project/enter.go](../../../server/router/project/enter.go)
- [server/api/v1/enter.go](../../../server/api/v1/enter.go)、[server/router/enter.go](../../../server/router/enter.go)
