# Service 示例

## 这个文件负责什么

`server/service/<模块>/` 的 Service 是无状态值，承载业务规则、数据访问与审计，不依赖 `gin.Context`。

## 什么时候应该这样写

- 新增查询、创建、更新、删除逻辑
- 需要在 API 之外复用的业务能力

## 推荐写法示例

```go
package order

import (
	"context"
	"errors"

	"optimus/server/global"
	commonResponse "optimus/server/model/common/response"
	orderModel "optimus/server/model/order"
	orderRequest "optimus/server/model/order/request"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderService struct{}

func (s *OrderService) List(ctx context.Context, projectIDs []uint, search orderRequest.OrderSearch) (commonResponse.PageResult[orderModel.Order], error) {
	search.Normalize()
	db := global.OPTIMUS_DB.WithContext(ctx).Model(&orderModel.Order{}).Where("project_id IN ?", projectIDs)
	if search.Status != nil {
		db = db.Where("status = ?", *search.Status)
	}
	if search.Keyword != nil {
		db = db.Where("name LIKE ?", "%"+*search.Keyword+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return commonResponse.PageResult[orderModel.Order]{}, err
	}
	list := make([]orderModel.Order, 0)
	err := db.Order("id DESC").Limit(search.PageSize).Offset((search.Page - 1) * search.PageSize).Find(&list).Error
	return commonResponse.PageResult[orderModel.Order]{List: list, Total: total, Page: search.Page, PageSize: search.PageSize}, err
}
```

## 为什么这样写

- Service 是**无字段的结构体**：数据库、配置和外部客户端在调用时从 `global` 或 `utils`（如 `utils.NewJWT()`、`utils/ldap.NewLDAP()`）获取，不在启动时注入；`initialize` 的契约测试会拒绝带字段的 Service。
- 第一个参数是 `context.Context`，用 `WithContext(ctx)` 透传；空集合返回 `[]` 而不是 `null`。
- 业务错误定义为包级 `Err*` 变量，由 API 层映射为 `FailWithMessage`。
- 需要审计的写操作在 Service 内记录（system 模块的 `recordAudit`），不依赖路由中间件。
- 服务间调用使用包级 `XxxServiceApp`（如 `cloudService.CloudAccountServiceApp`）。

## 常见错误

- 给 Service 加字段或 `NewXxxService` 构造函数
- 在 Service 里读写 `gin.Context`
- 忘记项目范围过滤（`project_id IN ?`），越权读取其他项目数据
- 列表返回 `nil` 切片导致 JSON 出现 `null`

## 真实参考文件

- [server/service/project/service.go](../../../server/service/project/service.go)
- [server/service/cloud/asset.go](../../../server/service/cloud/asset.go)（分页查询）
- [server/service/system/role.go](../../../server/service/system/role.go)（审计）
