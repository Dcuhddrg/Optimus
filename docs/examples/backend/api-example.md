# API 示例

## 这个文件负责什么

`server/api/v1/<模块>/` 处理 HTTP：绑定请求、取得当前用户与项目范围、调用 Service、按 GVA 协议输出。

## 什么时候应该这样写

- 新增 HTTP 接口
- 调整请求绑定或响应结构

## 推荐写法示例

```go
package order

import (
	"errors"

	"optimus/server/global"
	"optimus/server/model/common/response"
	orderRequest "optimus/server/model/order/request"
	orderService "optimus/server/service/order"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

type OrderApi struct{}

// ListOrders GET /orders
func (h *OrderApi) ListOrders(c *gin.Context) {
	if global.OPTIMUS_DB == nil {
		response.FailWithMessage("order service is not available", c)
		return
	}
	var search orderRequest.OrderSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		response.FailWithMessage("invalid order filter", c)
		return
	}
	scope := projectScope(c)
	page, err := orderSvc.List(c.Request.Context(), scope.AccessibleProjectIDs, search)
	if err != nil {
		response.FailWithMessage("failed to list orders", c)
		return
	}
	response.OkWithData(page, c)
}

// CreateOrder POST /orders
func (h *OrderApi) CreateOrder(c *gin.Context) {
	var req orderRequest.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid order payload", c)
		return
	}
	order, err := orderSvc.Create(c.Request.Context(), utils.GetUserID(c), req)
	if err != nil {
		if errors.Is(err, orderService.ErrOrderNotFound) {
			response.FailWithMessage(err.Error(), c)
			return
		}
		response.FailWithMessage("failed to create order", c)
		return
	}
	response.OkWithData(order, c)
}

// projectScope reads what the ProjectScope middleware stored for this request.
func projectScope(c *gin.Context) utils.ProjectScope {
	if raw, ok := c.Get(utils.ProjectScopeContextKey); ok {
		if scope, ok := raw.(utils.ProjectScope); ok {
			return scope
		}
	}
	return utils.ProjectScope{}
}
```

## 为什么这样写

- GET 用 `ShouldBindQuery`，JSON 请求体用 `ShouldBindJSON`。
- 只用 `response` 包输出：成功 `OkWithData`/`Ok`，业务失败 `FailWithMessage`（HTTP 200 + code 7）；不直接 `c.JSON`，也不返回 4xx/5xx。未认证由 JWT 中间件返回 401。
- 当前用户用 `utils.GetUserID(c)`（GVA 同名函数）；项目范围由 `ProjectScope` 中间件写入 `utils.ProjectScopeContextKey`，在 API 包内用 `projectScope(c)` 读取（同 `api/v1/cloud/helpers.go`），只信任其中的 `AccessibleProjectIDs`，不接受客户端传入的项目列表。
- 数据库不可用时返回业务失败，不让请求 panic。
- 内部错误不回显细节，已知业务错误才回传 `err.Error()`。

## 常见错误

- 在 API 里写 SQL 或业务规则
- 返回 HTTP 403/404 等状态码表示业务失败
- 自己解析 claims 或 `X-Project-Code`
- 把内部错误原文返回给前端

## 真实参考文件

- [server/api/v1/project/project.go](../../../server/api/v1/project/project.go)
- [server/api/v1/cloud/asset.go](../../../server/api/v1/cloud/asset.go)（分页）、[helpers.go](../../../server/api/v1/cloud/helpers.go)（项目范围）
- [server/model/common/response/response.go](../../../server/model/common/response/response.go)
