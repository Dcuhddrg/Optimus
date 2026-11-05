# 开发工作流

本文说明开发任务的推进顺序；协作职责、授权及验收边界以 [AGENT.MD](../../AGENT.MD) 为准。

## 开发顺序

1. 按 [文档入口](../README.md) 读取当前 Issue、必要评论和相关设计，明确需求范围、现有行为、模块边界及验收条件。
2. 涉及跨端时先核对 [接口契约](../frontend-backend/boundary.md)、调用方、权限及项目范围；破坏性变更明确发布与回退。
3. 按下文“分层实现顺序”实现，遵循 [后端分层规则](../modules/backend-layer-rules.md) 和现有模块模式。
4. 按影响范围验证，依据 [文档维护约定](documentation-workflow.md) 更新相关设计，并在 Issue/PR 记录交付与验证证据；业务验收按根规则执行。

## 分层实现顺序

参照 gin-vue-admin 的开发顺序，按 Optimus 的目录落地；各层写法见[开发示例](../examples/README.md)。

1. **Model**：`server/model/<module>/` 定义实体；`request/` 定义请求 DTO（分页查询内嵌 `PageInfo`），`response/` 定义响应 DTO。
2. **Service**：`server/service/<module>/` 实现业务规则与数据访问，无状态，在模块 `enter.go` 的 ServiceGroup 中登记。
3. **API**：`server/api/v1/<module>/` 按参数真实来源绑定、调用 Service、用 `response` 包输出，补齐 Swagger 注释。
4. **Router**：`server/router/<module>/` 的职责文件注册路径和中间件，由 `initialize/router.go` 挂到公开、私有或项目范围组。
5. **初始化与权限**：表迁移和种子数据放 `server/source/<module>/`；需要 RBAC 的菜单和 API 在 `service/system/seed.go` 的 `DefaultSystemPermissions` 登记权限码（`menu` 或 `api`，API 含方法和去前缀路径），由种子写入并同步 Casbin 策略。定时任务放 `server/task/`。
6. **前端**：`web/src/types/<module>.ts` 声明类型，`web/src/api/<module>.ts` 封装请求，`web/src/view/<module>/` 实现页面并接入 `web/src/router/` 与菜单权限码。
7. **联调与验证**：以后端真实接口为准，覆盖成功、业务失败、未认证、无权限、空数据和项目隔离；按 [AGENT.MD](../../AGENT.MD#验证) 的基线运行测试和构建。

前端可在后端接口稳定前按[接口契约](../frontend-backend/boundary.md)的 data 形状并行开发，联调时以后端实现为准。

## 工程操作与集成

- 代码发现优先使用 codebase-memory；工具、索引与验证步骤见 [工程操作参考](engineering-workflow.md)。
- 使用任务分支，按项目实际 PR 和 main 保护规则集成；不要求上游项目的 develop、feature 或 hotfix 分支体系。
- 提交信息清楚描述变更；仓库现有提交同时使用中文说明和语义化前缀，不强制虚构统一提交格式。
