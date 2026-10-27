# 架构决策索引

最近整理：2026-10-27。此处索引决定、依据和适用范围；规范方向确认不代表存量实现已经迁移或通过验收。

## 当前依据

| 主题 | 原始依据 | 适用范围与核查说明 |
| --- | --- | --- |
| 新旧模块的分层、结构迁移与装配 | [ADR-002](ADR-002-backend-structure-migration.md)、[当前架构约定](../current-architecture.md) | 统一 Router → API → Service → Model，存量 usecase/infra 必须分批迁移；方向已确认，实现及验收在 #11 跟进 |
| 账号同步、凭证加密、资产软删除、Provider 注册 | [账号资产设计的 Architecture Decisions](../../superpowers/specs/2026-05-13-cloud-account-asset-registry-design.md) | 原文标为 Approved；凭证与同步的实际安全性尚未在本轮验证 |
| 项目角色、资源归属与请求项目上下文 | [项目隔离设计的 Architecture Decisions](../../superpowers/specs/2026-05-14-multi-project-cloud-isolation-design.md) | 复用原文理由；保留 Open Questions，不能推断全部已解决 |
| 工单完成由任务完成驱动 | [任务管理设计](../../superpowers/specs/2026-06-02-ops-center-tasks-design.md) | 取代早期门户设计中的定时自动 done；已抽查核心联动源码，未做完整运行验收 |
| 多级审批链与当前审批人展示 | [审批链设计](../../superpowers/specs/2026-06-04-ticket-approval-chain-design.md) | 已有设计与计划，部分实现；逐级审批目标尚不能替代当前引擎事实 |

历史 spec / plan 保留当时的业务决定和交付证据；其分层及“保留旧结构”的约定不再产生当前豁免，现行要求以 ADR-002 为准。

## 已取代的历史记录

- [ADR-001](archive/ADR-001-backend-layering.md)：已被 ADR-002 取代，仅在追溯历史时读取。原路径保留短入口兼容旧链接，不纳入当前规范。

## 以后如何记录重大决定

涉及模块职责、跨模块数据模型、关键接口或安全边界变更时，使用未占用的稳定 ADR 编号记录：问题背景、约束、备选方案、选择理由、代价、批准依据、实施关联和替代关系。编号不复用，普通局部实现无需单独建立 ADR。

未决定的方案写“待确认”，不补造批准人和时间。新 ADR 取代旧决定时，两边互相链接并注明影响范围；源码漂移登记 [GitHub Issues](https://github.com/Dcuhddrg/Optimus/issues)，不要静默重写历史理由。
