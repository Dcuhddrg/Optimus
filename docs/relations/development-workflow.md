# 开发工作流

以 [AGENT.MD](../../AGENT.MD) 为规则入口，按任务读取 [文档入口](../README.md) 和关联 Issue、设计与能力状态。

## 开发顺序

1. 明确需求范围、现有行为、模块边界及验收条件。
2. 涉及跨端时先核对 [接口契约](../frontend-backend/boundary.md)、调用方、权限及项目范围；破坏性变更明确发布与回退。
3. 按 [后端分层规则](../modules/backend-layer-rules.md) 和现有模块模式组织 Model、Service、API、Router 及初始化；前端通过模块 API 接入。
4. 按影响范围验证并记录结果；事项的记录与验收遵循 [文档维护约定](documentation-workflow.md)。

## 代码和文档协作

- 代码发现优先使用 codebase-memory；工具、索引与验证步骤见 [工程操作参考](engineering-workflow.md)。
- Issue 与文档的维护、历史状态和证据规则见 [文档维护约定](documentation-workflow.md)。
- 使用任务分支，按项目实际 PR 和 main 保护规则集成；不要求上游项目的 develop、feature 或 hotfix 分支体系。
- 提交信息清楚描述变更；仓库现有提交同时使用中文说明和语义化前缀，不强制虚构统一提交格式。
