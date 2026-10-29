# Optimus 项目文档入口

最近架构规范整理：2026-10-27。能力核查范围和证据见状态表，文档整理不代表代码迁移或业务验收。

项目规则唯一正文：[AGENT.MD](../AGENT.MD)。`.claude/project_rules.md` 和 `.codex/project_rules.md` 仅为工具兼容适配层，负责指向统一规则和按需读取项目上下文，不保存独立规则副本。

## 从哪里开始

| 你要了解的内容          | 阅读入口                                                        | 维护职责                      |
| ---------------- | ----------------------------------------------------------- | ------------------------- |
| 长期方向、阶段依赖        | [平台路线图](memory/platform-roadmap.md)         | 用户确定目标与优先级，助手维护依据         |
| 当前实现到哪、哪些仍是 mock | [能力状态](project-status.md)                                   | 助手按代码和验证结果更新              |
| 统一架构与存量迁移要求       | [架构约定](relations/repo-profile.md)                | 遵循根规则；结构迁移已验收，详情见 #11 |
| 前后端协作与接口契约       | [接口契约](frontend-backend/boundary.md)                    | #7 文档先行；当前行为、目标规范与待确认项分开说明 |
| 当初为什么这样设计        | [决策索引](memory/decisions/README.md)                    | 维护原始理由和后续替代关系             |
| 问题、反馈和待决策事项      | [GitHub Issues](https://github.com/Dcuhddrg/Optimus/issues) | 用户决定取舍，助手持续跟进             |
| 文档状态、冲突与更新规则     | [文档维护约定](relations/documentation-workflow.md)                       | 每次工作均遵守                   |
| 图谱工具用法、运行与条件验证步骤 | [工程操作参考](relations/engineering-workflow.md)                         | 按任务触发读取                   |

后端统一采用 `Router -> API -> Service -> Model`。project、cloud、system 的结构迁移已通过 [#11](https://github.com/Dcuhddrg/Optimus/issues/11) 验收；路由与依赖装配的现状见 [后端分层规则](modules/backend-layer-rules.md)，接口响应迁移由 #18 单独跟进。

## 按主题读取工程上下文

参考 gin-vue-admin 的组织方式，按任务选择入口，不需要通读所有目录：

- [relations](relations/README.md)：系统结构、技术栈及运行边界。
- [modules](modules/README.md)：模块职责、后端分层与迁移规则。
- [frontend-backend](frontend-backend/README.md)：前后端契约、前端规范与复用。
- [examples](examples/README.md)：基于已验证实现的开发示例入口。
- [memory](memory/README.md)：长期规则、能力证据与 GitHub Issue 入口。

架构决策维护在 [memory/decisions](memory/decisions/README.md)，产品设计、计划和路线图保留各自职责。旧架构入口已删除；现行系统架构和后端分层分别从 relations 与 modules 读取。

## 功能设计与计划

- 云资源：[基础设计](superpowers/specs/2026-05-13-cloud-resource-foundation-design.md)、[账号与资产](superpowers/specs/2026-05-13-cloud-account-asset-registry-design.md)、[项目隔离](superpowers/specs/2026-05-14-multi-project-cloud-isolation-design.md)。
- 运维中心：[服务门户](superpowers/specs/2026-05-28-service-portal-design.md)、[任务管理](superpowers/specs/2026-06-02-ops-center-tasks-design.md)、[审批链](superpowers/specs/2026-06-04-ticket-approval-chain-design.md)。
- 对应实施步骤在 [plans](superpowers/plans/)；历史设计在 [archive](superpowers/archive/)。目录归档位置不等于功能已验证。

## 开始和结束一项工作

开始时读取项目规则和本入口，再按任务读取相关能力行、Issue 正文/必要评论及适用设计；不通读全部文档或全部 Issues。代码发现优先走 codebase-memory，并确认索引状态；图谱不是测试或验收证据。

结束时记录实际改动、验证结果、剩余问题及相关编号。涉及行为或边界演进时更新设计关联。需要用户决策的事项保留待评估，不由助手补造结论。

新问题统一在 GitHub 提出，进度和讨论在对应 Issue 维护；docs 保存产品、架构和能力证据。本地不再维护问题台账或编号映射正文。聊天中形成的重要结论须写入对应 Issue 或设计文档，才能作为长期依据。
