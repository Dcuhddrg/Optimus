# 开发示例

## 读取建议

### 后端示例

按 [Request](backend/request-example.md) → [Model](backend/model-example.md) → [Service](backend/service-example.md) → [API](backend/api-example.md) → [Router](backend/router-example.md) → [enter.go](backend/enter-go-example.md) 阅读；插件入口参考 [plugin.go](backend/plugin-go-example.md)。

### 前端示例

按 [API](frontend/api-example.md) → [Pinia 状态管理](frontend/pinia-example.md) → [View](frontend/view-example.md) → [工具复用](frontend/utils-usage-example.md) 阅读。

### 独立插件示例

[插件示例与边界](plugin/full-plugin-example.md) 是设计参照，不代表完整插件框架已存在。

## 使用边界

- backend 与 frontend 的 13 份文档保留 gin-vue-admin 原始模板；包路径、业务、Vue/Pinia 和插件 SDK 等内容及末尾参考文件指向上游，不代表 Optimus 已有实现或依赖，也不按项目现状反向改写模板。
- Optimus 使用 Go/Gin 与 React/TypeScript。现行约束见 [AGENT.MD](../../AGENT.MD)、[后端分层](../modules/backend-layer-rules.md)、[接口契约](../frontend-backend/boundary.md) 和[前端规范](../frontend-backend/frontend-rules.md)；使用示例前核对项目实际能力和调用方。
- [#45](https://github.com/Dcuhddrg/Optimus/issues/45) 跟踪模板核对及迁移，后端聚合已按 [#48](https://github.com/Dcuhddrg/Optimus/pull/48) 落地；前端和插件适用范围、业务验收及 #18 响应协议迁移分别跟踪。
