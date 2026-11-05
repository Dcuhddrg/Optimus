# 插件示例

> **状态：插件机制尚未实现。** Optimus 当前没有 GVA 的插件 SDK（`utils/plugin/v2`）、自动注册或插件市场；本篇说明在现有结构下插件应如何组织，独立注册或自动加载需要先设计确认（见 [AGENT.MD](../../../AGENT.MD#插件与模块)）。

## 这个文件负责什么

插件是放在 `server/plugin/<name>/` 下的自包含功能，内部沿用 `Router → API → Service → Model` 分层，通过一个入口函数接入主程序。

## 什么时候应该这样写

- 功能可整体启用或移除，且不属于 system/project/cloud 任一模块
- 已就插件边界和注册方式完成设计确认

## 推荐写法示例

```text
server/plugin/announcement/
├── api/          # ApiGroup + 包级 Service 别名
├── model/        # 模型与 request/response
├── router/       # 职责子组路由
├── service/      # 无状态 Service
├── source/       # 种子数据（RegisterInit）
└── plugin.go     # 入口
```

```go
// server/plugin/announcement/plugin.go
package announcement

import (
	"optimus/server/plugin/announcement/router"

	"github.com/gin-gonic/gin"
)

// Register wires the plugin into the private route group. It is called
// explicitly from initialize.Routers; Optimus has no plugin auto-loading.
func Register(privateGroup *gin.RouterGroup) {
	router.RouterGroupApp.InitAnnouncementRouter(privateGroup)
}
```

## 为什么这样写

- 插件内部写法与主模块一致，复用 JWT、Casbin、ProjectScope 和 GVA 响应协议，不另建鉴权或响应格式。
- 入口由 `initialize.Routers` 显式调用，启动顺序清楚；表与种子数据通过插件 `source/` 的 `RegisterInit` 接入启动初始化。
- 插件新增的 API 同样要登记到权限清单，否则会被 Casbin 拒绝。

## 常见错误

- 照搬 GVA 的 `interfaces.Register` / 插件 SDK 调用（Optimus 不存在）
- 插件绕过主程序的鉴权或响应协议
- 未经设计确认就引入自动加载、插件市场等机制

## 真实参考文件

- [server/initialize/router.go](../../../server/initialize/router.go)（显式注册位置）
- [server/source/project/project.go](../../../server/source/project/project.go)（`RegisterInit` 写法）
- [插件示例与边界](../plugin/full-plugin-example.md)
