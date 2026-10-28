# 项目文档治理初版 Implementation Plan

> 历史实施记录：本计划描述仓库台账初版。后续问题跟踪已按用户指令迁移至 GitHub，当前规则见 [维护约定](../../relations/documentation-workflow.md)，原 DOC-001 见 [GitHub #1](https://github.com/Dcuhddrg/Optimus/issues/1)。以下清单和校验数字保留当时记录，不用于恢复本地问题状态。

> 执行方式：按 superpowers:executing-plans 在当前会话逐项实施。用户已确认上一轮评估建议并授权初版实现。

**Goal:** 让项目方向、架构约束、能力状态和问题反馈可查、可追溯，避免把设计误当成已交付事实。

**Architecture:** 复用现有 docs，增加入口、能力状态、文档治理约定、问题台账和决策索引。问题状态以仓库台账为唯一来源；将来迁移外部 Issue 时保留编号与链接，不双写。

**Tech Stack:** Markdown、Git、codebase-memory；不修改应用代码，不增加依赖。

**状态：** 执行完待验证（2026-10-26 文档校验通过；初版约定待用户审阅，关联 DOC-001）。

## 范围与验收

- 文档和 .gitignore、.claude/project_rules.md；保留已有未提交的审批链测试。
- 用户负责产品与架构取舍，助手负责核查、维护和实现；不虚构批准记录或业务验收结果。
- 所有新增文档有入口，已有历史约束保留并注明演进关系。
- 新增 Markdown 不再被 docs/superpowers 忽略，保留其他忽略规则。
- 未核实能力标为待核实，代码存在与测试通过、用户验收分别记录。
- 相对文件链接可解析，git diff --check 通过，变更限于文档治理。

## Task 1：建立入口和维护约定

- [x] 新建 docs/README.md、docs/governance/documentation.md。
- [x] 在根 README 和项目规则中链接文档入口，修正 README 与已存在 pnpm 约定冲突的命令。
- [x] 移除 .gitignore 中 docs/superpowers/ 的整体忽略。

## Task 2：记录现状、问题和决策

- [x] 新建 docs/project-status.md，记录核查日期、提交基线、证据强度及待验证项。
- [x] 新建 docs/issues/README.md，定义稳定编号、状态、优先级、责任分工、验收和 Issue 迁移规则，录入已确认文档问题和产品待决策项。
- [x] 新建 docs/architecture/decisions/README.md，索引现有决定；不冒充新的 ADR 批准。

## Task 3：修正状态与演进关系

- 修改文件及边界：
  - `docs/roadmaps/finops-devops-platform-roadmap.md`：将旧基线标为历史，当前快照引用能力表，不重新决定阶段优先级。
  - `docs/architecture/current-architecture.md`：用能力表链接替代重复进度表，保留架构契约。
  - `docs/superpowers/specs/2026-05-28-service-portal-design.md`：顶部增加局部替代和历史背景说明，不删除原始范围。
  - `docs/superpowers/specs/2026-06-02-ops-center-tasks-design.md`：链接已有计划，标明源码抽查不等于交付验收。
  - `docs/superpowers/specs/2026-06-04-ticket-approval-chain-design.md`：链接已有计划及问题，区分局部实现与设计目标。
  - `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`：顶部标明空章节及承接候选，并修正归档导致的失效引用，不补造决定。
- [x] 更新路线图当前状态快照，架构文档引用统一能力状态，保留历史阶段说明。
- [x] 标注服务门户、任务管理与审批链设计的局部替代关系，区分设计批准与交付验收。
- [x] 在云资源基础设计顶部标明空章节和参考文档，保留原文并登记跟进。

## Task 4：验证与交付

- 字段和允许值：由 `docs/governance/documentation.md` 定义文档生命周期、实现/验证分离、责任和复核触发条件；由 `docs/issues/README.md` 定义编号、优先级、状态、来源、责任、关联、验收、下一步和验证记录。
- 校验范围：上述六份新文档（含本计划）及本轮修改的现有 Markdown；检查本地文件及目录链接、显式使用的标题锚点。外部 URL 不发网络请求，代码示例中的占位链接不计入。
- 执行命令：`python3 /tmp/optimus-docs-governance-check.py`（本轮生成的标准库校验脚本，记录本次目标清单并解析本地 Markdown 链接及标题，核对台账编号唯一、表格状态和关键字段），预期退出 0，报告目标文档数、链接数及台账数；脚本不是应用依赖。
- Git 检查：`git diff --check` 预期退出 0；`git check-ignore --no-index docs/superpowers/specs/2026-06-04-ticket-approval-chain-design.md docs/superpowers/plans/2026-10-26-documentation-governance.md` 预期退出 1 且无输出（不再忽略）。
- 边界检查：`git status --short --untracked-files=all` 和 `git diff --stat`；应用源码无本轮修改，原有未提交测试仍存在。
- [x] 用 Python 校验本次文档中的相对文件链接、台账编号与关键状态字段。
- [x] 用 git check-ignore 核实 Markdown 可被 Git 发现；git diff --check、git status 核查边界。
- [x] 更新本计划完成项和校验结果；本次不声称业务测试或架构全面验收通过，不自动提交或推送。

## 校验记录

- 2026-10-26：计划经过独立审阅并按意见补充路径、状态字段及校验步骤，复审批准实施。
- `python3 /tmp/optimus-docs-governance-check.py`：通过，14 份文档、85 个本地链接/锚点、5 个唯一台账编号；状态/优先级/必填字段和 Git 可见性检查通过。脚本为本次会话临时校验工具，不是仓库长期依赖；后续可按本节定义的范围重建。
- `git diff --check`：通过；原有未提交的审批链测试保持不动，本轮变更限于文档及对应规则。
- 不代表业务功能验证通过；DOC-001 保持待验证，等待用户审阅初版。
