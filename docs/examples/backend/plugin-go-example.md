# plugin.go 示例

> 改写自 [gin-vue-admin / backend/plugin-go-example.md](https://github.com/flipped-aurora/gin-vue-admin/blob/cee04b08d1262036ca33e3026e1e13d29471c7d5/aiDoc/examples/backend/plugin-go-example.md)（`cee04b08`）。2026-10-30 适配 Optimus：保留上游讲解结构，代码、路径及现状说明已修改；上游许可见 [LICENSE.gin-vue-admin](../LICENSE.gin-vue-admin)。

## 这个文件负责什么

插件入口应只协调依赖构造、初始化和路由接入，业务仍按既有分层组织。目录约定是 `server/plugin/<name>/` 与 `web/src/plugin/<name>/`。

## 什么时候应该这样写

- 已确认能力需要独立插件边界。
- 准备明确插件生命周期及装配方式。
- 普通业务模块与插件的职责需要比较。

## 推荐写法示例

本页只给装配关系示意，不声明存在可直接导入的插件注册 SDK：

```text
启动装配 → 显式构造插件依赖 → 插件 API handlers → 已有路由组注册
                                     ↓
                               Service / Model
```

上游 `interfaces.Register(Plugin)`、`plugin/v2` 和隐式 init 注册未作为 Optimus 的已核实能力。当前没有已验证的完整插件样板可在此作为标准复制；先核对或设计入口，再编写实现。仅普通 CRUD 不应为符合示例而强行插件化。

## 为什么这样写

上游入口保持薄的组织方式可沿用，具体注册接口和初始化能力必须以本项目事实为准。新增生命周期或自动发现属于设计工作，不由示例暗中引入。

## 常见错误

- 写出不存在的插件 SDK 导入。
- 声明菜单、字典或数据库初始化会自动完成。
- 为示例创建未经批准的业务模块。

## 真实参考文件

- [真实模块构造参照](../../../server/initialize/system_module.go)、[路由参照](../../../server/initialize/router.go)；这两者不是完整插件实现。
- [完整插件边界](../plugin/full-plugin-example.md)、[根规则插件约定](../../../AGENT.MD#插件与模块)。
