# 完整插件示例

> 参照 gin-vue-admin `cee04b08` 的 `aiDoc/examples/plugin/full-plugin-example.md`，按 Optimus 写法改写。`order` 为占位业务；注册框架尚未实现，见 [ADR-004](../../adr/0004-gva-plugin-registration.md)。

## 这个文件负责什么

说明一个能力以插件方式存在时，前后端目录、初始化入口和接口组织如何落位。

## 推荐目录结构

```text
server/plugin/order/
├── api/
│   ├── enter.go          ApiGroup 与 Service 别名
│   └── order.go
├── initialize/
│   ├── permission.go     菜单与 API 权限码，同步 Casbin
│   ├── gorm.go           表迁移
│   ├── router.go         公开/私有组与中间件
│   └── viper.go          按需：插件配置段
├── model/
│   ├── order.go
│   └── request/order.go
├── router/
│   ├── enter.go          RouterGroup 与 API 别名
│   └── order.go
├── service/
│   ├── enter.go          ServiceGroup
│   └── order.go
└── plugin.go             导出 Plugin，init() 登记，Register 调度

server/plugin/register.go   空白导入 optimus/server/plugin/order

web/src/plugin/order/
├── api/
│   └── order.ts
├── components/
└── view/
    ├── OrderListPage.tsx
    └── OrderListPage.module.css
```

## 推荐实现顺序

1. 先定义 model 和 request
2. 再写 service
3. 再写 api 和 router
4. 再补 `initialize/`（权限、表、路由）
5. 再补 `plugin.go`，并在 `register.go` 加空白导入
6. 最后接前端 `api/`、`view/`、路由与菜单权限码

## 为什么这样写

- 前后端插件结构对称，更容易被 AI 和开发者理解
- 初始化逻辑集中在 `initialize/`，入口文件保持薄
- `plugin.go` 只做登记与调度，不承担业务细节

## 常见错误

- 插件目录缺少 `enter.go`，聚合方式与主系统不一致
- 后端插件已拆层，前端插件却把请求、状态和展示塞进一个页面
- 插件路由组漏挂 JWT、Casbin 或 ProjectScope 中间件
- 复制 GVA 的 `initialize.Dictionary`、`.vue`、Pinia 或 `config.js`（Optimus 无对应能力）

## 真实参考文件

- 后端分层与装配：[server/router/system/enter.go](../../../server/router/system/enter.go)、[server/service/system/enter.go](../../../server/service/system/enter.go)
- 路由组与中间件：[server/initialize/router.go](../../../server/initialize/router.go)
- 前端页面写法：[web/src/view/project/ProjectsPage.tsx](../../../web/src/view/project/ProjectsPage.tsx)
- 上游：gin-vue-admin `cee04b08` 的 `server/plugin/announcement/`
