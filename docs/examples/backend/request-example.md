# Request 示例

## 这个文件负责什么

`server/model/<模块>/request/` 的结构体承接前端入参、查询条件和分页参数，不负责数据库关联或业务逻辑。

## 什么时候应该这样写

- 新增列表查询参数
- 新增创建 / 更新接口的请求体
- 需要复用分页字段时

## 推荐写法示例

```go
package request

import commonRequest "optimus/server/model/common/request"

type CreateOrderRequest struct {
	Name   string `json:"name" form:"name" binding:"required"`
	Status string `json:"status" form:"status"`
	Remark string `json:"remark" form:"remark"`
}

// UpdateOrderRequest uses pointers so an omitted field is left unchanged.
type UpdateOrderRequest struct {
	Name   *string `json:"name" form:"name"`
	Remark *string `json:"remark" form:"remark"`
}

type OrderSearch struct {
	commonRequest.PageInfo
	Status  *string `json:"status" form:"status"`
	Keyword *string `json:"keyword" form:"keyword"`
}
```

## 为什么这样写

- `json` 与 `form` 标签一起写，JSON 请求体和查询参数都能绑定；必填字段用 `binding:"required"`。
- 分页查询内嵌公共 `PageInfo`（page 从 1 开始，默认 20、上限 200，由 `Normalize()` 处理）。
- 可选过滤与局部更新用指针，区分“未传”和“传了空值”。
- 请求结构和数据库模型分离，不把 model 当请求体。

## 常见错误

- 在 request 里放数据库关联或内部字段（如 `ProjectID` 由服务端从项目范围取得）
- 漏写 `form` 标签，查询参数绑定不全
- 直接复用 model 作为请求体

## 真实参考文件

- [server/model/cloud/request/asset.go](../../../server/model/cloud/request/asset.go)
- [server/model/project/request/project.go](../../../server/model/project/request/project.go)
- [server/model/common/request/page_info.go](../../../server/model/common/request/page_info.go)
