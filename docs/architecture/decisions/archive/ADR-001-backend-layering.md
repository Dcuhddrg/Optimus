# ADR-001：统一后端分层规范与存量实现的适用边界

> 状态：已被 [ADR-002](../ADR-002-backend-structure-migration.md) 取代，仅供历史追溯；不得作为当前开发或迁移豁免依据。

- 整理日期：2026-10-27。
- 规范方向：已确认，依据用户明确要求与根目录 [AGENT.MD](../../../../AGENT.MD)；本轮文档交付的审阅与验收在 [Issue #6](https://github.com/Dcuhddrg/Optimus/issues/6) 跟进。
- 适用范围：后端逻辑分层、注册装配和旧规范的替代关系；不包含业务代码迁移。

## 背景

原架构文档把 usecase/infra 组织方式写为所有新模块的强制规则。用户随后明确以 gin-vue-admin 的 Router → API → Service → Model 为工程基准，根规则已经确认，架构文档需要消除双重约束。

## 决定及理由

- 统一采用 `Router -> API -> Service -> Model`，API 负责 HTTP，Service 不依赖 `gin.Context`。
- 保留 `enter.go` 分组注册与组合入口，启动和依赖装配沿用现有 `main` / `initialize` 流程。
- 旧 usecase/infra 内容保留为历史方案和存量实现说明，不再规定新模块必须创建这些目录。
- 保留其历史理由：通过接口分离业务与持久化，支持纯业务单元测试和外部适配。但这些理由不再形成另一套全仓库强制分层。
- 分开记录规范与实际实现，既满足统一框架协作规则，也避免未经核查和授权迁移存量模块。

各层职责及依赖方向只在 [当前架构](../../current-architecture.md#backend-分层契约) 展开，本记录保存决定与沿革。

## 存量情况与影响

本轮图谱仅确认 system、project、cloud 服务中仍有 usecase/infra 目录，没有实施代码迁移、运行验证或业务验收。存量模块内原有 Store、Provider、初始化注入和持久化封装继续按任务复用，不能据此宣布所有代码已经符合统一分层。

历史 spec / plan 中涉及旧分层的内容按其原阶段理解；与当前根规则冲突的分层要求已被本决定取代，历史业务决定不因此一并作废。任何代码迁移、接口改变及依赖调整都需要单独收敛范围并验证。

## 旧方案原文（已被取代，仅供追溯）

以下从整理前的 `docs/architecture/current-architecture.md` 保留。引文内“强制”“必须”只描述旧方案，不是当前开发要求。

> ## Backend 分层契约
>
> ```
> main → initialize → router → api/handlers → usecase/service → infra/persistence
> ```
>
> **强制约束：**
>
> - `usecase/` 只依赖标准库，定义 Store / Provider 等接口，**不感知 GORM / Gin / 任何框架**
> - `infra/persistence/` 实现 Store 接口、持有 `*gorm.DB`
> - `infra/<adapter>/` 实现外部协议适配器（如 LDAP、JWT、云厂商 SDK）
> - `initialize/<module>_module.go` 负责依赖注入：GORM 实现 → usecase → handlers
> - handler 不得直接调用 GORM；usecase 不得 import infra
>
> **新模块必须遵循的目录约定：**
>
> ```
> server/
> ├── api/v1/<module>/         # HTTP handlers
> ├── router/<module>/         # 路由注册（public / private 分组）
> ├── service/<module>/
> │   ├── usecase/             # 接口 + 业务逻辑，可纯单元测试
> │   └── infra/
> │       ├── persistence/     # GORM Store 实现
> │       └── provider/        # 外部 SDK 适配器
> ├── model/<module>/          # GORM 模型
> └── initialize/<module>.go   # 模块组装
> ```
