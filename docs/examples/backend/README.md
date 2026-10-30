# 后端示例

> 改写自 [gin-vue-admin / backend/README.md](https://github.com/flipped-aurora/gin-vue-admin/blob/cee04b08d1262036ca33e3026e1e13d29471c7d5/aiDoc/examples/backend/README.md)（`cee04b08`）。2026-10-30 适配 Optimus：保留上游讲解结构，代码、路径及现状说明已修改；上游许可见 [LICENSE.gin-vue-admin](../LICENSE.gin-vue-admin)。

## 包含内容

- [Request](request-example.md)：请求来源、DTO 与查询条件。
- [Model](model-example.md)：实体与持久化字段。
- [Service](service-example.md)：业务、数据访问协作与测试。
- [API](api-example.md)：绑定、HTTP 状态与 Swagger。
- [Router](router-example.md)：分组、权限和 handler 绑定。
- [enter.go](enter-go-example.md)：路由汇总与构造注入。
- [plugin.go](plugin-go-example.md)：插件入口的适用边界。

## 适用范围

新增或维护 Go 模块时按职责选择阅读；不要按旧示例重新引入 usecase/infra。完整规则见 [后端分层](../../modules/backend-layer-rules.md)。
