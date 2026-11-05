# 插件开发约束

> 参照 gin-vue-admin `cee04b08` 的 `aiDoc/modules/plugin-development.md`，按 Optimus 现状改写。项目规则以 [AGENT.MD](../../AGENT.MD#插件与模块) 为准。

**现状：** Optimus 目前没有 `server/plugin/` 和 `web/src/plugin/` 目录，也没有 GVA 的插件 SDK（`utils/plugin/v2`）、自动注册或插件市场。本文是新建插件时的约束；独立注册或自动加载须先设计确认。

## 什么时候做成插件

- 功能可以整体启用或移除，且不属于 system、project、cloud 任一现有模块。
- 已就插件边界、数据模型、权限和装配方式完成设计确认。

不满足以上条件时，按普通模块放进 `server/{api/v1,service,model,router}/<module>/` 与 `web/src/{api,view}/`。

## 后端插件结构

```text
server/plugin/<name>/
├── api/             HTTP 输入绑定、调用 Service、组织响应
├── model/           实体
├── model/request/   请求 DTO，分页查询内嵌公共 PageInfo
├── router/          路由分组，enter.go 聚合子 Router
├── service/         无状态 Service，业务与持久化
├── initialize/      表迁移、路由注册、种子数据等初始化
└── plugin.go        插件入口
```

插件内部同样遵守 `Router → API → Service → Model` 和[后端分层规则](backend-layer-rules.md)：`enter.go` 只做聚合，Service 不依赖 `gin.Context`，响应只用 `model/common/response`。

## 前端插件结构

```text
web/src/plugin/<name>/
├── api/             模块函数，统一走 utils/request 的 apiRequest
├── components/      仅插件内复用的组件
└── view/            React/TypeScript 页面
```

前端插件复用 `web/src/utils/`、`web/src/auth/`、`web/src/locale/` 和 `web/src/components/` 的已有能力，见[前端工具函数约束](../frontend-backend/frontend-utils.md)。

## 插件入口职责

> **状态：提议，待确认。** 插件的注册与装配方式尚未设计确认（见 [AGENT.MD](../../AGENT.MD#插件与模块)）。下面是起草时的候选方案，确认前不作为实现标准；确认后改写本节并去掉此标注。

GVA 的 `plugin.go` 实现插件接口、在 `init()` 中注册、通过 `Register` 挂载路由、通过 `RouterPath` 返回根路径。Optimus 没有对应接口，候选方案是 `plugin.go`：

- 提供一个显式入口函数，由 `initialize` 调用，不依赖包级 `init()` 的隐式副作用；
- 接收 initialize 创建的公开/私有路由组，挂载插件路由，JWT、RBAC 和项目范围中间件与主系统一致；
- 返回或声明插件根路径，避免与已有路由冲突；
- 注册表迁移和种子数据（沿用 `source/` 的 `RegisterInit` 方式），菜单与权限码随种子数据一起登记。

无论最终采用哪种注册方式，路由中间件、响应协议和项目隔离须与主系统一致，这一条不随方案变化。

## 设计原则

- 自包含：插件代码只放在自己的目录下，主系统只通过入口函数接入。
- 可配置：配置项进 `config.yaml` 的独立段，缺省时插件不影响主系统启动。
- 一致：命名、响应协议、权限和项目隔离与主系统相同，不另建一套。
- 克制：只为已确认的扩展点预留接口，不提前做插件市场或热加载。

## 开发顺序

1. 明确插件边界与数据模型，完成设计确认。
2. 完成后端 Model、Request、Service、API、Router 和 initialize。
3. 完成前端 API 封装、页面和组件。
4. 补齐菜单、权限码、联调和测试，按 [AGENT.MD](../../AGENT.MD#验证) 的基线验证。

示例见[完整插件示例](../examples/plugin/full-plugin-example.md)和 [plugin.go 示例](../examples/backend/plugin-go-example.md)。
