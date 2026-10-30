# Optimus 文档入口

先读 [AGENT.MD](../AGENT.MD)，再按任务选择相关文档，不必通读整个目录。规则冲突时以根规则和用户当前指令为准。

## 工程上下文

| 目录 | 内容 |
| --- | --- |
| [relations](relations/README.md) | 项目画像、系统关系、开发流程和工程操作 |
| [modules](modules/README.md) | 模块职责与后端分层 |
| [frontend-backend](frontend-backend/README.md) | 前后端协作、接口契约与前端规范 |
| [examples](examples/README.md) | 后端、React 前端和插件边界的开发示例；沿用 gin-vue-admin 讲解结构并适配本项目 |
| [memory](memory/README.md) | 项目长期方向与重要决策 |

## 产品与实施

- [平台路线图](memory/long-term/platform-roadmap.md)：长期方向与阶段依赖。
- [设计](superpowers/specs/)：[云资源基础](superpowers/specs/2026-05-13-cloud-resource-foundation-design.md)、[账号资产](superpowers/specs/2026-05-13-cloud-account-asset-registry-design.md)、[项目隔离](superpowers/specs/2026-05-14-multi-project-cloud-isolation-design.md)、[服务门户](superpowers/specs/2026-05-28-service-portal-design.md)、[任务管理](superpowers/specs/2026-06-02-ops-center-tasks-design.md)和[审批链](superpowers/specs/2026-06-04-ticket-approval-chain-design.md)。
- [实施计划索引](superpowers/plans/README.md)：计划范围与按需读取说明；历史记录见 [归档](superpowers/archive/)。
- [GitHub Issues](https://github.com/Dcuhddrg/Optimus/issues)：需求、反馈、问题、优先级、进度及验证/验收证据的唯一跟踪入口；交付关联提交与 PR。

文档和 Issue 的维护、状态及证据规则见 [文档维护约定](relations/documentation-workflow.md)。
