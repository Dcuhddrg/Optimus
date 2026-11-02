# 开发示例

## 读取建议

### 后端示例

按 [Request](backend/request-example.md) → [Model](backend/model-example.md) → [Service](backend/service-example.md) → [API](backend/api-example.md) → [Router](backend/router-example.md) → [enter.go](backend/enter-go-example.md) 阅读；插件入口参考 [plugin.go](backend/plugin-go-example.md)。

### 前端示例

按 [API](frontend/api-example.md) → [Pinia 状态管理](frontend/pinia-example.md) → [View](frontend/view-example.md) → [工具复用](frontend/utils-usage-example.md) 阅读。

### 独立插件示例

[完整插件边界](plugin/full-plugin-example.md) 暂保留原有内容，不属于本次前后端模板替换范围。

## 使用边界

- backend 与 frontend 的 13 份文档为 gin-vue-admin 原始参考模板；示例中的包路径、订单业务、全局聚合、插件 SDK、Vue/Pinia 和 Element Plus 不代表 Optimus 已有实现或依赖。
- 这些模板面向全局规范与后续迁移，示例业务不限定适用模块。不得把当前实现反向改写成模板标准，也不得把文档替换视为源码整改完成。
- Optimus 当前仍使用 Go/Gin 与 React/TypeScript。模板中的前端框架和后端装配差异，需在后续规范与迁移方案中明确；此次替换不更换技术栈，不新增模板中的业务或插件框架。
- 模板末尾的真实参考文件指向参考框架的目录与文件，不是 Optimus 的本地可用文件清单。迁移前须重新核对本项目符号、调用方和能力边界。
- 现行项目约束见 [AGENT.MD](../../AGENT.MD)、[后端分层](../modules/backend-layer-rules.md)、[接口契约](../frontend-backend/boundary.md) 和 [前端规范](../frontend-backend/frontend-rules.md)。模板与当前规则的差异在 [#45](https://github.com/Dcuhddrg/Optimus/issues/45) 的后续阶段确认，不在文档替换阶段擅自作架构取舍。
- #45 记录固定模板版本、文件核对、验证和分阶段交付；源码迁移、业务验收及 #18 接口协议迁移分别跟踪。
