# Model 示例

## 这个文件负责什么

`server/model/<模块>/` 定义数据库实体与持久化字段，是 Service 与数据库交互的基础。

## 什么时候应该这样写

- 新增一张业务表
- 为现有表补字段
- 定义 GORM 字段约束

## 推荐写法示例

```go
package order

import (
	"optimus/server/global"

	"gorm.io/gorm"
)

type Order struct {
	global.OPTIMUS_MODEL
	ProjectID uint           `gorm:"not null;index" json:"projectId"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	Status    string         `gorm:"size:16;not null;default:'active'" json:"status"`
	Remark    string         `gorm:"size:512" json:"remark"`
	CreatorID uint           `gorm:"not null;index" json:"creatorId"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Order) TableName() string {
	return "orders"
}
```

## 为什么这样写

- 内嵌 `global.OPTIMUS_MODEL`（`id`、`createdAt`、`updatedAt`），对应 GVA 的 `GVA_MODEL`。
- `OPTIMUS_MODEL` **不含软删除**（与 GVA 不同）；需要软删除的表自行声明 `DeletedAt` 并隐藏于 JSON。
- 显式写 `TableName()`，表名不随命名策略变化。
- 新表要加入模块的 `AutoMigrate*`，由 `source/<模块>` 初始化器在启动时迁移。

## 常见错误

- 缺少 `json` 或 `gorm` 标签
- 把仅用于请求或展示的字段写进 model
- 前后端对同一字段用不同类型（尤其 `status`、ID、时间）
- 新增 model 却没加入 `AutoMigrate*`

## 真实参考文件

- [server/model/project/project.go](../../../server/model/project/project.go)
- [server/global/model.go](../../../server/global/model.go)
- [server/service/project/project.go](../../../server/service/project/project.go)（`AutoMigrateProject`）
