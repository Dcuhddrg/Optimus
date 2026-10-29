# 当前能力状态

- 最近后端核查：2026-10-27；前端 mock 等未复验条目保留此前核查范围。
- 核查基线：`12c3e9f`；本轮验证包含路由与授权边界整改（[#12](https://github.com/Dcuhddrg/Optimus/issues/12)、[#13](https://github.com/Dcuhddrg/Optimus/issues/13)），对应交付提交和 PR 以 Issue 中的关联记录为准。
- 核查方式：codebase-memory 定向审计、后端全量测试及构建。隔离 MySQL 基线 149 个测试通过；整改后 190 个测试（含子测试）通过、0 个用例跳过，中间件及装配竞态检查通过。未进行前端构建、真实云联调、部署或端到端验收；本表不是生产就绪声明。
- 各阶段长期目标见 [路线图](memory/long-term/platform-roadmap.md)，具体问题状态见 [GitHub Issues](https://github.com/Dcuhddrg/Optimus/issues)。
- 架构目标以 [ADR-002](memory/long-term/decisions/ADR-002-backend-structure-migration.md) 为准：存量后端结构须分批迁移；project、cloud、system 的结构改动已合入，三批和 #11 已由用户验收，结构迁移已关闭。原 190 个测试仅证明首批边界修复的基线，迁移验证见 #11。

## 能力与证据

| 能力 | 实现状态 | 当前依据及限制 | 验证状态 / 下一步 |
| --- | --- | --- | --- |
| 系统身份、角色、权限、审计 | 已接后端；业务验收待核实 | [架构约定](relations/repo-profile.md)；后端单元/持久化测试通过，系统入口与项目解析故障的解耦已补回归 | 测试基线见 [#2](https://github.com/Dcuhddrg/Optimus/issues/2)；外部认证、真实会话及审计页面等仍需按场景验收 |
| 云账号、资产、项目隔离 | 已接后端；真实云联调待核实 | [账号资产设计](superpowers/specs/2026-05-13-cloud-account-asset-registry-design.md)、[隔离计划](superpowers/plans/2026-05-16-sp2-cloud-project-binding.md)；隔离库持久化、成员授权和新增路由隔离回归通过 | 证据见 [#2](https://github.com/Dcuhddrg/Optimus/issues/2)、[#11](https://github.com/Dcuhddrg/Optimus/issues/11)；真实云同步、部署行为和产品验收未覆盖 |
| 服务门户与工单 | 前端 mock，完整性待验证 | [服务门户设计](superpowers/specs/2026-05-28-service-portal-design.md)；[mock 引擎](../web/src/view/service/data/ticketMockEngine.ts) 中 schedule 仍以定时器推进 submitted→approving→processing | 本轮仅源码抽查；未验收完整 UI、刷新及存储异常行为 |
| 运维任务与工单联动 | 前端 mock，核心联动已见源码 | [taskStore](../web/src/view/service/data/taskStore.ts) 的 complete 将任务置 done 并调用 ticketStore.advance(processing, done)；[引擎](../web/src/view/service/data/ticketMockEngine.ts) 已无自动 done 定时器；提交 `193e026` 有关联展示改动 | 源码抽查，非运行验证；完整链路按 [#2](https://github.com/Dcuhddrg/Optimus/issues/2) 补验 |
| 多级审批链展示 | 部分实现，尚未验证整链 | [类型定义](../web/src/view/service/ticketTypes.ts) 与提交 `4b713f1` 已加 ApprovalStep 和事件字段；引擎仍是固定单步进入 processing；[目标设计](superpowers/specs/2026-06-04-ticket-approval-chain-design.md) 不等于已实现 | 按 [#4](https://github.com/Dcuhddrg/Optimus/issues/4) 跟进 |
| 后端真实工单、任务与人工审批 | 现有运维 spec 列为后续范围；独立需求待评估 | [任务范围](superpowers/specs/2026-06-02-ops-center-tasks-design.md)、[审批范围](superpowers/specs/2026-06-04-ticket-approval-chain-design.md)；本轮未做全仓不存在性证明 | 不能按前端 mock 宣称有真实执行能力；关联 [#5](https://github.com/Dcuhddrg/Optimus/issues/5) |
| FinOps、自动化与策略执行 | 路线图规划，具体实现待核实 | [路线图](memory/long-term/platform-roadmap.md)；未核查相关实现 | 未确认需求与交付边界，不宣称已交付 |

## 更新方式

只更新有新增证据的行，保留日期和基线。源码存在、测试通过、业务验收三个层次分别记录。无法确认时保留待核实，不根据文件名、归档目录或计划勾选推断完成。

初版文档校验记录在 [#1](https://github.com/Dcuhddrg/Optimus/issues/1)；本轮后端基线和首批整改记录在 [#2](https://github.com/Dcuhddrg/Optimus/issues/2)、[#11](https://github.com/Dcuhddrg/Optimus/issues/11)。响应、分页和错误字段尚未迁移到统一 envelope，契约已由 [#7](https://github.com/Dcuhddrg/Optimus/issues/7) 确认，代码迁移在 [#18](https://github.com/Dcuhddrg/Optimus/issues/18) 跟进，不能据测试通过宣称迁移已完成。
