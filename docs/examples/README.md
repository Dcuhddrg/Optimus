# 开发示例

## 读取建议

### 后端示例

按 [Request](backend/request-example.md) → [Model](backend/model-example.md) → [Service](backend/service-example.md) → [API](backend/api-example.md) → [Router](backend/router-example.md) → [enter.go](backend/enter-go-example.md) 阅读；插件入口参考 [plugin.go](backend/plugin-go-example.md)。

### 前端示例

按 [API](frontend/api-example.md) → [共享状态（Context）](frontend/context-example.md) → [View](frontend/view-example.md) → [公共能力复用](frontend/reuse-example.md) 阅读。

### 独立插件示例

[插件示例与边界](plugin/full-plugin-example.md) 是设计参照，不代表完整插件框架已存在。

## 使用边界

- backend 的 8 份与 frontend 的 5 份文档按用户决定（2026-11-05）参照 gin-vue-admin `cee04b08` 的结构改写为 Optimus 写法（Go/Gin 包路径与 GVA 协议、React/TypeScript），写法与真实参考文件对应 Optimus 源码；`order` 为占位业务；插件注册框架已确定照 GVA v2（[ADR-004](../adr/0004-gva-plugin-registration.md)），代码尚未实现。
- Optimus 使用 Go/Gin 与 React/TypeScript。现行约束见 [AGENT.MD](../../AGENT.MD)、[后端分层](../modules/backend-layer-rules.md)、[接口契约](../frontend-backend/boundary.md) 和[前端规范](../frontend-backend/frontend-rules.md)；使用示例前核对项目实际能力和调用方。
- 模板核对与迁移记录见 [#45](https://github.com/Dcuhddrg/Optimus/issues/45)。
