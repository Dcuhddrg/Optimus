# 实施计划索引

计划是特定阶段的实施记录。读取前先看 [AGENT.MD](../../../AGENT.MD)、关联 Issue 及其交付/验收记录；涉及旧代码路径时以当前代码与规范为准。保留在本目录不等于已排期，也不表示步骤仍可直接执行；逐项交付依据、验收缺口及保留原因见 [#10 核对记录](https://github.com/Dcuhddrg/Optimus/issues/10)。

## 云资源与项目

- [云账号与资产](2026-05-13-cloud-account-asset-registry.md)、[云资源基础设计计划](2026-05-13-cloud-resource-foundation-spec.md)。
- [项目后端 SP1](2026-05-15-sp1-project-backend.md)、[云资源项目隔离 SP2](2026-05-16-sp2-cloud-project-binding.md)、[项目管理前端 SP3](2026-05-16-sp3-project-frontend-shell.md)。
- [云资源前端和标签 SP4](2026-05-17-sp4-cloud-frontend-and-tags.md)、[资产页补全](2026-05-17-sp4-assets-page-completion.md)。

## 运维中心

- [服务门户](2026-05-28-service-portal.md)、[卡片尺寸](2026-05-31-service-card-sizing.md)、[门户头部](2026-05-31-service-portal-header.md)。
- [任务管理](2026-06-02-ops-center-tasks.md)、[多级审批链](2026-06-04-ticket-approval-chain.md)。审批链尚未完成，见 [#4](https://github.com/Dcuhddrg/Optimus/issues/4)。

## 系统治理

- [审计查询与阅读体验](2026-10-29-audit-query-experience.md)：分页、筛选、现有字段详情及 CSV 导出，关联 [#28](https://github.com/Dcuhddrg/Optimus/issues/28)。
- [审计记录质量](2026-10-29-audit-record-quality.md)：关键事件摘要、未认证账号线索及写入失败可观测性，关联 [#36](https://github.com/Dcuhddrg/Optimus/issues/36)。

## 状态维护

状态与归档遵循 [文档维护约定](../../relations/documentation-workflow.md#状态与事实)；现存计划的清理在 [#10](https://github.com/Dcuhddrg/Optimus/issues/10) 跟进。云资源基础规格相关核查按用户决定暂停，见 [#3](https://github.com/Dcuhddrg/Optimus/issues/3)；其余未取得验收依据的计划保留，不由文件日期或提交记录推定验收。

历史计划见 [archive/plans](../archive/plans/)。已验收的 [会话恢复计划](../archive/plans/2026-10-29-session-recovery.md) 对应 [#27](https://github.com/Dcuhddrg/Optimus/issues/27) 与 [PR #32](https://github.com/Dcuhddrg/Optimus/pull/32)；归档不扩大其验证范围，后续补验另见 [#34](https://github.com/Dcuhddrg/Optimus/issues/34)。
