# Optimus 文档入口

先读 [AGENT.MD](../AGENT.MD)，再按任务选择相关文档，不必通读整个目录。规则冲突时以根规则和用户当前指令为准。

## 工程上下文

| 目录 | 内容 |
| --- | --- |
| [relations](relations/README.md) | 项目画像、系统关系、开发流程和工程操作 |
| [modules](modules/README.md) | 模块职责与后端分层 |
| [frontend-backend](frontend-backend/README.md) | 前后端协作、接口契约与前端规范 |
| [examples](examples/README.md) | 前后端参考示例及插件边界 |
| [memory](memory/README.md) | 项目长期方向与业务上下文 |
| [adr](adr/README.md) | 架构决策记录（ADR） |

## 按任务入口

| 任务 | 先读 |
| --- | --- |
| 新增或修改后端接口 | [开发工作流](relations/development-workflow.md#分层实现顺序)、[后端分层规则](modules/backend-layer-rules.md)、[接口契约](frontend-backend/boundary.md)、[后端示例](examples/README.md#后端示例) |
| 新增或修改前端页面 | [前端规范](frontend-backend/frontend-rules.md)、[前端工具函数](frontend-backend/frontend-utils.md)、[接口契约](frontend-backend/boundary.md)、[前端示例](examples/README.md#前端示例) |
| 新增模块或插件 | [模块说明索引](modules/module-index.md)、[插件开发约束](modules/plugin-development.md)、[系统关系图](relations/system-map.md) |
| 权限、认证或项目隔离 | [接口契约第 6 节](frontend-backend/boundary.md#6-请求认证与权限协作)、[项目隔离设计](design/2026-05-14-multi-project-cloud-isolation-design.md) |
| 记录重要决定 | [ADR 索引](adr/README.md) |
| 运行、部署与验证 | [项目画像](relations/repo-profile.md)、[工程操作参考](relations/engineering-workflow.md) |
| 更新文档或 Issue | [文档维护约定](relations/documentation-workflow.md) |

## 产品与实施

- [平台路线图](memory/long-term/platform-roadmap.md)：长期方向与阶段依赖。
- [设计](design/README.md)：[云资源基础](design/2026-05-13-cloud-resource-foundation-design.md)、[账号资产](design/2026-05-13-cloud-account-asset-registry-design.md)、[项目隔离](design/2026-05-14-multi-project-cloud-isolation-design.md)、[服务门户](design/2026-05-28-service-portal-design.md)、[任务管理](design/2026-06-02-ops-center-tasks-design.md)和[审批链](design/archive/2026-06-04-ticket-approval-chain-design.md)。
- 设计文档目录 `design/` 只存在本地，已加入 `.gitignore`，不推送到 GitHub，新克隆里没有这些文件。上面的链接只在本机保留了 `docs/design/`，或从 Git 历史恢复了它时可用（例如 `git restore --source=20e511d --worktree -- docs/design` 恢复移出前的版本，只写工作区、不会重新跟踪）；在 GitHub 上和新克隆里都会失效。
- 实施计划：新计划和任务记在 GitHub Issue；仍在进行的计划放在 `design/plans/`（目前没有），交付后删除。已交付的设计见 [design/archive](design/archive/)。
- [GitHub Issues](https://github.com/Dcuhddrg/Optimus/issues)：需求、反馈、问题、优先级、进度及验证/验收证据的唯一跟踪入口；交付关联提交与 PR。

文档和 Issue 的维护、状态及证据规则见 [文档维护约定](relations/documentation-workflow.md)。
