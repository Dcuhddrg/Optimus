# 后端示例

讲解型示例，按 gin-vue-admin `cee04b08` 的 `aiDoc/examples/backend` 结构改写为 Optimus 写法（包路径 `optimus/server/...`、GVA 响应协议、无状态 Service、职责子组路由）。示例中的 `order` 模块是占位业务，不代表已实现；写法以每篇末尾的真实参考文件为准。

## 包含内容

| 示例 | 说明 |
| --- | --- |
| [request-example.md](request-example.md) | 请求结构：JSON/查询绑定、分页 |
| [model-example.md](model-example.md) | 数据模型：`OPTIMUS_MODEL`、表名、按需软删除 |
| [service-example.md](service-example.md) | 无状态业务服务：`global.OPTIMUS_DB`、错误、审计 |
| [api-example.md](api-example.md) | API：绑定、调用服务、GVA 响应 |
| [router-example.md](router-example.md) | 职责子组路由与注册位置 |
| [enter-go-example.md](enter-go-example.md) | 根级与模块级 `enter.go` 聚合及别名 |
| [plugin-go-example.md](plugin-go-example.md) | 插件目录与接入方式（插件机制待设计确认） |

## 适用范围

新增或修改以下文件前先阅读对应示例：

- `server/model/**`
- `server/service/**`
- `server/api/**`
- `server/router/**`
- `server/source/**`、`server/task/**`（见[后端分层](../../modules/backend-layer-rules.md)）
- `server/plugin/**`
