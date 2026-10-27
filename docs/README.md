# Optimus 项目文档入口

最近架构规范整理：2026-10-27。能力核查范围和证据见状态表，文档整理不代表代码迁移或业务验收。

项目规则唯一正文：[AGENT.MD](../AGENT.MD)。`.claude/project_rules.md` 和 `.codex/project_rules.md` 仅为工具兼容适配层，负责指向统一规则和按需读取项目上下文，不保存独立规则副本。

## 从哪里开始

| 你要了解的内容 | 阅读入口 | 维护职责 |
| --- | --- | --- |
| 长期方向、阶段依赖 | [平台路线图](roadmaps/finops-devops-platform-roadmap.md) | 用户确定目标与优先级，助手维护依据 |
| 当前实现到哪、哪些仍是 mock | [能力状态](project-status.md) | 助手按代码和验证结果更新 |
| 当前架构规范与存量实现说明 | [架构约定](architecture/current-architecture.md) | 遵循根规则，区分规范与实现；重要边界变更由用户决定 |
| 当初为什么这样设计 | [决策索引](architecture/decisions/README.md) | 维护原始理由和后续替代关系 |
| 问题、反馈和待决策事项 | [GitHub Issues](https://github.com/Dcuhddrg/Optimus/issues)；[旧编号映射](issues/README.md) | 用户决定取舍，助手持续跟进 |
| 文档状态、冲突与更新规则 | [文档维护约定](governance/documentation.md) | 每次工作均遵守 |
| 图谱工具用法、运行与条件验证步骤 | [工程操作参考](governance/engineering.md) | 按任务触发读取 |

后端统一规范为 `Router -> API -> Service -> Model`，保留 `enter.go` 注册和组合入口。旧 `usecase/infra` 方案的适用边界和替代关系见 [ADR-001](architecture/decisions/ADR-001-backend-layering.md)，不据此自动迁移存量代码。

## 功能设计与计划

- 云资源：[基础设计](superpowers/specs/2026-05-13-cloud-resource-foundation-design.md)、[账号与资产](superpowers/specs/2026-05-13-cloud-account-asset-registry-design.md)、[项目隔离](superpowers/specs/2026-05-14-multi-project-cloud-isolation-design.md)。
- 运维中心：[服务门户](superpowers/specs/2026-05-28-service-portal-design.md)、[任务管理](superpowers/specs/2026-06-02-ops-center-tasks-design.md)、[审批链](superpowers/specs/2026-06-04-ticket-approval-chain-design.md)。
- 对应实施步骤在 [plans](superpowers/plans/)；历史设计在 [archive](superpowers/archive/)。目录归档位置不等于功能已验证。

## 开始和结束一项工作

开始时读取项目规则和本入口，再按任务读取相关能力行、Issue 正文/必要评论及适用设计；不通读全部文档或全部 Issues。代码发现优先走 codebase-memory，并确认索引状态；图谱不是测试或验收证据。

结束时记录实际改动、验证结果、剩余问题及相关编号。涉及行为或边界演进时更新设计关联。需要用户决策的事项保留待评估，不由助手补造结论。

新问题统一在 GitHub 提出，进度和讨论在对应 Issue 维护；docs 保存产品、架构和能力证据。原有 5 条已迁移，本地仅保留旧编号映射。聊天中形成的重要结论须写入对应 Issue 或设计文档，才能作为长期依据。
