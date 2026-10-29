# 开发工作流

本文说明开发任务的推进顺序；协作职责、授权及验收边界以 [AGENT.MD](../../AGENT.MD) 为准。

## 开发顺序

1. 按 [文档入口](../README.md) 读取当前 Issue、必要评论、相关能力行和适用设计，明确需求范围、现有行为、模块边界及验收条件。
2. 涉及跨端时先核对 [接口契约](../frontend-backend/boundary.md)、调用方、权限及项目范围；破坏性变更明确发布与回退。
3. 按 [后端分层规则](../modules/backend-layer-rules.md) 和现有模块模式组织 Model、Service、API、Router 及初始化；前端通过模块 API 接入。
4. 按影响范围验证，依据 [文档维护约定](documentation-workflow.md) 更新相关设计，并在 Issue/PR 记录交付与验证证据；业务验收按根规则执行。

## 工程操作与集成

- 代码发现优先使用 codebase-memory；工具、索引与验证步骤见 [工程操作参考](engineering-workflow.md)。
- 使用任务分支，按项目实际 PR 和 main 保护规则集成；不要求上游项目的 develop、feature 或 hotfix 分支体系。
- 提交信息清楚描述变更；仓库现有提交同时使用中文说明和语义化前缀，不强制虚构统一提交格式。
