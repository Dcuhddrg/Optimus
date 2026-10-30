# 开发示例

## 读取建议

### 新增后端能力时

按 [Request](backend/request-example.md) → [Model](backend/model-example.md) → [Service](backend/service-example.md) → [API](backend/api-example.md) → [Router](backend/router-example.md) → [enter.go](backend/enter-go-example.md) 阅读。

### 新增前端能力时

按 [API](frontend/api-example.md) → [状态管理](frontend/state-example.md) → [View](frontend/view-example.md) → [工具复用](frontend/utils-usage-example.md) 阅读。状态管理示例使用 React Context。

### 新增插件时

先读 [完整插件边界](plugin/full-plugin-example.md) 与 [plugin.go](backend/plugin-go-example.md)，分清目录约定与尚未建立的运行机制。

## 使用原则

- 优先参考下面链接的真实文件，再参考教学片段；片段说明所在函数/文件上下文，不是可整目录复制的新业务模块。
- 规范来源仍是 [AGENT.MD](../../AGENT.MD)、[后端分层](../modules/backend-layer-rules.md)、[接口契约](../frontend-backend/boundary.md) 和 [前端规则](../frontend-backend/frontend-rules.md)，示例不维护第二套规则。
- 本组主要以系统审计查询为贯穿实例：Model → Store/Service → API → Router/initialize → 前端 API → 页面与语言资源；它已包含真实分页和 CSV，不用上游订单业务代替项目事实。
- 审计查询的本地迁移已使用 `{code,data:{list,total,page,pageSize},msg}`，模块 API 显式解包为分页 data；其他旧接口仍返回原始 DTO。已迁移范围及交付以 [#18](https://github.com/Dcuhddrg/Optimus/issues/18) 为准，CSV 成功不包 JSON。
- 上游的 `GVA_MODEL`、`global.GVA_DB`、全局 ServiceGroupApp、Vue/Pinia、Element Plus 和插件注册 SDK 均不自动成为本仓库现有能力。
- 示例随对应接口变更同步；审计协议首批迁移见 [#18](https://github.com/Dcuhddrg/Optimus/issues/18)，原审计查询及记录链路的实现与验证分别见 [#28](https://github.com/Dcuhddrg/Optimus/issues/28)、[#36](https://github.com/Dcuhddrg/Optimus/issues/36)。路径、差异和本轮核查限制见 [#8](https://github.com/Dcuhddrg/Optimus/issues/8)，源码变更后通过 codebase-memory-mcp 重新核对，不把教学片段当作完整实现。

## 验证清单

- 请求来源、字段、分页、HTTP 错误及响应类型与真实调用一致，当前 DTO 和 #18 目标协议分开说明。
- 新增能力纳入现有构造注入及 enter.go 路由链，核对权限种子、服务端授权和必要的项目范围，不创建平行注册入口。
- 前端同步模块 API、类型、路由、导航与中英文文案，检查加载、空、失败/重试及权限不足状态；菜单可见性不代替 API 权限。
- 覆盖正常、失败、越权和受影响的回归；新增表或插件初始化还需确认生命周期与环境条件，不能据文件存在推断可运行。
- 按 [工程操作参考](../relations/engineering-workflow.md#验证步骤) 和 [根规则验证要求](../../AGENT.MD#验证) 运行最小必要检查，将命令、实际结果、基线、跳过/阻塞及未覆盖项记录到对应 Issue/PR。
- 仅修改文档时检查文件、章节、来源、引用与片段语法；语法检查不等于片段可独立构建，既有测试结果不冒充本次业务验收。
