# plugin.go 示例

> **状态：注册方式已确定（[ADR-004](../../adr/0004-gva-plugin-registration.md)），框架代码尚未实现。** 下文的 `optimus/server/utils/plugin/v2` 和 `server/plugin/announcement` 是随第一个插件落地的目标写法，当前仓库中不存在。

## 这个文件负责什么

插件入口负责把插件登记到注册表，并在系统启动时按顺序完成权限登记、表迁移和路由注册。

## 什么时候应该这样写

- 新增一个 Optimus 插件
- 需要把插件初始化逻辑收敛到统一入口

## 推荐写法示例

```go
// server/plugin/announcement/plugin.go
package announcement

import (
	"context"

	"optimus/server/plugin/announcement/initialize"
	interfaces "optimus/server/utils/plugin/v2"

	"github.com/gin-gonic/gin"
)

var _ interfaces.Plugin = (*plugin)(nil)

var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

func (p *plugin) Register(engine *gin.Engine) {
	ctx := context.Background()
	initialize.Permission(ctx) // 菜单与 API 权限码，同步 Casbin
	initialize.Gorm(ctx)       // 表迁移
	initialize.Router(engine)  // 公开/私有组与职责路由
}
```

```go
// server/plugin/announcement/initialize/router.go
package initialize

import (
	"optimus/server/middleware"
	"optimus/server/plugin/announcement/router"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	public := engine.Group(utils.RouterPrefix())
	private := engine.Group(utils.RouterPrefix())
	private.Use(middleware.JWTAuth(), middleware.CasbinHandler())
	router.RouterGroupApp.InitInfoRouter(public, private)
}
```

```go
// server/plugin/register.go
package plugin

import (
	_ "optimus/server/plugin/announcement"
)
```

## 为什么这样写

- `init()` 中登记，`register.go` 的空白导入让启动时能发现插件，与 gin-vue-admin v2 一致。
- `Register` 把初始化步骤按顺序集中起来，细节下沉到插件的 `initialize/`。
- 插件自建路由组时挂上与主系统相同的中间件，鉴权和响应协议不另建一套。

## 常见错误

- 在 `plugin.go` 里塞业务细节
- 忘记调用 `interfaces.Register`，或忘记在 `register.go` 加空白导入
- 不加别名直接导入 `utils/plugin/v2`：其包名是 `plugin`，与 `type plugin struct{}` 重名，编译失败
- 私有组漏挂 `JWTAuth`/`CasbinHandler`，访问云资源的路由漏挂 `ProjectScope`
- 新增 API 没有登记权限码，请求被 Casbin 拒绝

## 真实参考文件

- [server/initialize/router.go](../../../server/initialize/router.go)（主系统的路由组与中间件）
- [server/service/system/seed.go](../../../server/service/system/seed.go)（权限码格式）
- [server/source/project/project.go](../../../server/source/project/project.go)（表迁移与种子写法）
- 上游：gin-vue-admin `cee04b08` 的 `server/plugin/announcement/plugin.go`
