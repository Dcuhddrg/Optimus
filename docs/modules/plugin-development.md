# 插件开发约束

> 参照 gin-vue-admin `cee04b08` 的 `aiDoc/modules/plugin-development.md`，按 Optimus 现状改写。项目规则以 [AGENT.MD](../../AGENT.MD#插件与模块) 为准，注册方式的决定见 [ADR-004](../adr/0004-gva-plugin-registration.md)。

**现状：** 注册方式已确定照 gin-vue-admin v2，但框架代码（`server/utils/plugin/v2`、`server/plugin/register.go`、initialize 中的插件挂载）尚未实现，随第一个插件一起落地。目前没有 `server/plugin/` 和 `web/src/plugin/` 目录。

## 什么时候做成插件

- 功能可以整体启用或移除，且不属于 system、project、cloud 任一现有模块。
- 已就插件边界、数据模型和权限完成设计确认。

不满足以上条件时，按普通模块放进 `server/{api/v1,service,model,router}/<module>/` 与 `web/src/{api,view}/`。

## 后端插件结构

```text
server/plugin/<name>/
├── api/             enter.go 聚合 ApiGroup；HTTP 输入绑定、调用 Service、组织响应
├── model/           实体
├── model/request/   请求 DTO，分页查询内嵌公共 PageInfo
├── router/          enter.go 聚合子 Router；职责文件注册路径
├── service/         enter.go 聚合 ServiceGroup；无状态 Service
├── initialize/      permission.go、gorm.go、router.go，按需 viper.go
└── plugin.go        导出 Plugin，init() 中登记，Register 只做调度

server/plugin/register.go    空白导入所有插件
```

插件内部同样遵守 `Router → API → Service → Model` 和[后端分层规则](backend-layer-rules.md)：`enter.go` 只做聚合，Service 不依赖 `gin.Context`，响应只用 `model/common/response`。

## 前端插件结构

```text
web/src/plugin/<name>/
├── api/             模块函数，统一走 utils/request 的 apiRequest
├── components/      仅插件内复用的组件
└── view/            React/TypeScript 页面
```

前端插件复用 `web/src/utils/`、`web/src/auth/`、`web/src/locale/` 和 `web/src/components/` 的已有能力，见[前端工具函数约束](../frontend-backend/frontend-utils.md)。页面接入 `web/src/router/` 并使用插件登记的菜单权限码。

## 插件入口职责

与 gin-vue-admin v2 一致：

- `plugin.go` 导出 `var Plugin = new(plugin)`，用 `var _ plugin.Plugin = (*plugin)(nil)` 校验接口；在 `init()` 中调用 `plugin.Register(Plugin)`。
- `Register(engine *gin.Engine)` 按顺序调用 `initialize.Permission`、`initialize.Gorm`、`initialize.Router`，不写业务细节。
- 插件的 `initialize/router.go` 用 `utils.RouterPrefix()` 创建公开和私有组，私有组挂 `middleware.JWTAuth()`、`middleware.CasbinHandler()`；访问云资源的路由还要挂 `middleware.ProjectScope`。
- 插件的菜单与 API 权限码按主系统格式（`menu`/`api`，API 含方法和去前缀路径）登记并同步 Casbin，否则请求会被拒绝。
- 新插件必须在 `server/plugin/register.go` 加空白导入，否则不会生效。

GVA 的 `initialize.Menu`、`initialize.Api` 在 Optimus 合并为权限登记；GVA 的 `initialize.Dictionary` 无对应能力，不创建。

## 设计原则

- 自包含：插件代码只放在自己的目录下，主系统只通过注册表接入。
- 可配置：配置项进 `config.yaml` 的独立段，缺省时插件不影响主系统启动。
- 一致：命名、响应协议、权限和项目隔离与主系统相同，不另建一套。
- 克制：不做插件市场、热加载或运行时启停。

## 开发顺序

1. 明确插件边界与数据模型，完成设计确认。
2. 第一个插件落地时，先补框架：`server/utils/plugin/v2`、`server/plugin/register.go` 和 initialize 中的挂载。
3. 完成后端 Model、Request、Service、API、Router 和 `initialize/`，最后写 `plugin.go`。
4. 完成前端 API 封装、页面和组件。
5. 补齐权限码、联调和测试，按 [AGENT.MD](../../AGENT.MD#验证) 的基线验证。

示例见[完整插件示例](../examples/plugin/full-plugin-example.md)和 [plugin.go 示例](../examples/backend/plugin-go-example.md)。
