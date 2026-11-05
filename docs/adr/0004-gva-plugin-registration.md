# ADR-004：插件注册照 gin-vue-admin v2

- 日期：2026-11-05。
- 状态：已由用户确认；框架代码尚未实现，随第一个插件一起落地。跟踪见 [#68](https://github.com/Dcuhddrg/Optimus/issues/68)。
- 现行规则：[插件开发约束](../modules/plugin-development.md)。

## 背景

Optimus 没有插件机制，`docs/modules/plugin-development.md` 原先只给出“显式入口函数、由 initialize 调用”的候选方案。后端其余部分（分层、装配、响应协议、认证）都已对齐 gin-vue-admin（`cee04b08`）。

## 决定

插件注册与 gin-vue-admin v2 一致：

- `server/utils/plugin/v2` 定义 `Plugin` 接口（`Register(engine *gin.Engine)`）和注册表（`Register`、`Registered`）。
- 每个插件在 `server/plugin/<name>/plugin.go` 以别名 `interfaces` 导入注册包（避免与插件内的 `type plugin struct{}` 重名），导出 `Plugin` 值，在包的 `init()` 中调用 `interfaces.Register(Plugin)` 登记。
- `server/plugin/register.go` 用空白导入汇总所有插件；`initialize` 在主路由注册后遍历 `plugin.Registered()` 调用各插件的 `Register`。
- 插件的 `Register` 只做调度，按顺序调用插件内 `initialize/` 的权限登记、表迁移和路由注册。

## 考虑过的方案

- **显式列表**：initialize 手写插件列表，不用 `init()` 和注册表，更直观，但与 GVA 不一致。
- **暂不定**：等第一个插件出现再设计。

选择 GVA v2，是为了让插件和后端其他部分一样按上游结构组织，示例可以直接对照上游。

## 代价

- 注册依赖包级 `init()` 和空白导入，忘记在 `register.go` 加导入时插件不会生效，也不会报错。
- 插件在自己的 `initialize/router.go` 里创建公开、私有路由组，必须挂上与主系统相同的中间件（JWT、Casbin，云资源还要 ProjectScope），否则会放宽权限。
