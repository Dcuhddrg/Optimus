# Router 示例

## 这个文件负责什么

`server/router/<模块>/` 的职责 Router 在自己的 `InitXxxRouter` 内建立子组并绑定相对路径；鉴权中间件由 `initialize/router.go` 挂在公共/私有组上。

## 什么时候应该这样写

- 新增一组接口
- 给一组接口挂专属中间件

## 推荐写法示例

```go
package order

import "github.com/gin-gonic/gin"

type OrderRouter struct{}

func (r *OrderRouter) InitOrderRouter(privateGroup *gin.RouterGroup) {
	orderRouter := privateGroup.Group("orders")
	orderRouter.GET("", orderApi.ListOrders)
	orderRouter.POST("", orderApi.CreateOrder)
	orderRouter.GET(":id", orderApi.GetOrder)
	orderRouter.PUT(":id", orderApi.UpdateOrder)
	orderRouter.DELETE(":id", orderApi.DeleteOrder)
}
```

在 `server/initialize/router.go` 中注册：

```go
orderRouter := &routerGroup.Order
orderRouter.InitOrderRouter(privateGroup) // 需要项目范围时传 cloudGroup 这类挂了 ProjectScope 的组
```

## 为什么这样写

- 子组 + 相对路径，路由按职责聚在一个文件，前缀（默认 `/api/v1`）由 `initialize` 统一决定。
- 私有组已挂 `JWTAuth` 和 `CasbinHandler`；Casbin 策略按“去掉前缀的路径 + 方法”匹配，所以新增接口要在模块的权限清单（如 `DefaultProjectPermissions`）里登记 `Path: "/orders/:id"` 这类路径，由 `source/` 授予 `admin`。
- Optimus 不用 GVA 的 `OperationRecord` 中间件和 `RouterWithoutRecord` 分组，审计在 Service 内记录。
- 路径风格沿用 Optimus 的 REST（`GET/POST/PUT/DELETE` + 资源路径），不照搬 GVA 的 `getOrderList` 动词式路径。

## 常见错误

- 在共享组上直接写完整路径，不建子组
- 把公开接口注册到私有组，或反之
- 新增接口忘记在权限清单登记，导致所有角色都被 Casbin 拒绝
- 只写了 Router 却没在 `initialize/router.go` 调用 `InitXxxRouter`

## 真实参考文件

- [server/router/project/project.go](../../../server/router/project/project.go)
- [server/router/system/auth.go](../../../server/router/system/auth.go)（公开 + 私有子组）
- [server/initialize/router.go](../../../server/initialize/router.go)
