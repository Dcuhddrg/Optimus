# System Service Boundary Design

日期：2026-05-12

## 背景

在 [`2026-05-11-backend-architecture-gap-assessment.md`](./2026-05-11-backend-architecture-gap-assessment.md) 中，`问题 2` 被定义为：

`service` 混合用例逻辑与基础设施实现。

当前 `server/service/system` 的主要问题不是接口缺失，而是抽象只停留在类型层，没有真正落实到包边界：

- `UserService`、`RoleService`、`LDAPService`、`AuditService` 与对应的 `Gorm*Store` 实现位于同包，且大多位于同文件
- `AuthService` 虽然已经把 `GormAuthStore` 拆到独立文件，但仍与用例逻辑共处 `service/system` 包内
- `JWT` 签发/解析实现、LDAP 客户端实现也与业务能力同域混放
- `SystemModule` 已经收回装配权，但其内部依赖仍然指向“逻辑与实现混包”的结构

这导致当前结构只能算“显式接口 + 隐式边界”，还不是可以稳定演进的能力边界。

本设计只处理 `问题 2`，不与更大范围的目录治理或运行时治理问题并行推进。

## 目标

本设计的目标是把 `system` 域内部收紧成真实的用例层 / 基础设施层边界，同时保持当前外层工程壳子可延续。

成功标准：

- `system` 域内的业务用例与基础设施实现分处不同包边界
- `usecase` 层不再直接依赖 `gorm`、LDAP 客户端库、JWT 细节实现
- `infra` 层不再承载业务规则决策
- `initialize.SystemModule` 通过装配 `infra -> usecase` 提供能力，而不是直接依赖混合包
- `api` / `middleware` 的对外行为保持兼容，本轮不借拆包改变协议层语义

## 非目标

- 不重写外层 `initialize / router / api / service / global` 工程骨架
- 不在本轮中消除 `global`
- 不在本轮中处理启动流程副作用问题
- 不在本轮中把认证、授权、审计进一步抽成全项目级独立 runtime capability
- 不顺手调整路由路径、HTTP 响应结构、权限语义

## 方案选择

本轮采用“外层兼容、内层收紧”的方案：

- 保留现有外层工程壳子
- 在 `system` 域内部建立真实的 `usecase` / `infra` 包边界
- 继续由 `initialize.SystemModule` 统一装配

不采用“只拆文件、不拆包边界”的原因：

- 只能减轻阅读负担，不能真正解决“能力与实现没有隔离”的问题

不采用“直接引入全项目 `application / infrastructure` 顶层目录”的原因：

- 会把 `问题 2` 与全项目目录重命名、import 迁移、其他业务域跟进绑定在一起，当前改动成本过高

不采用“只选单一纵切面做样板”的原因：

- 更适合验证思路，不足以完成对 `问题 2` 的修复
- `auth`、`user`、`role`、`ldap`、`audit` 长期并存两套边界模型会继续制造结构债

## 设计概览

本轮在 `server/service/system/` 下建立新的域内分层，建议结构如下：

- `server/service/system/usecase`
  - `auth.go`
  - `user.go`
  - `role.go`
  - `ldap.go`
  - `audit.go`
  - `rbac.go`
- `server/service/system/infra/persistence`
  - `auth_store_gorm.go`
  - `user_store_gorm.go`
  - `role_store_gorm.go`
  - `ldap_store_gorm.go`
  - `audit_store_gorm.go`
  - `rbac_store_gorm.go`
- `server/service/system/infra/identity`
  - `jwt.go`
- `server/service/system/infra/directory`
  - `ldap_client.go`
- `server/service/system/shared`
  - 默认不建立；只有出现真实反向依赖时才引入，用于承载少量跨用例共享、且不带基础设施倾向的稳定类型

运行时装配流向固定为：

`initialize/SystemModule -> infra implementations -> usecase -> api/middleware`

静态包依赖方向固定为：

`api/middleware -> usecase contracts or compat seams`

`infra implementations -> usecase contracts`

反向依赖禁止：

- `usecase` 依赖 `gorm`
- `usecase` 直接 import LDAP 客户端实现或 JWT 库细节
- `api` / `middleware` 直接构造 store / client / signer
- `infra` 反向依赖 `api`

## 各层职责边界

### `usecase`

负责：

- 定义输入输出模型
- 编排一个业务能力所需的依赖
- 承载稳定业务规则与错误语义
- 定义它依赖的接口，如 `UserStore`、`TokenIssuer`、`LDAPAuthenticator`、`AuditWriter`

不负责：

- 持有 `gorm.DB`
- 表达 SQL / `Joins` / `Where`
- 直接调用 LDAP 客户端实现
- 直接使用 JWT 库
- 感知 Gin / HTTP

### `infra/persistence`

负责：

- `Gorm*Store` 持久化实现
- 查询语义、事务、关联、排序、去重
- 将数据库结果映射到 `usecase` 需要的数据形状

不负责：

- “最后一个 admin-capable 用户不能移除”这类业务决策
- HTTP 错误语义
- JWT / LDAP 外部能力

### `infra/identity`

负责：

- token 签发和解析的适配器实现
- 将当前 JWT 细节包裹在 `usecase` 可消费的接口之后

### `infra/directory`

负责：

- LDAP 连接、测试、认证的客户端适配器实现

### `shared`

负责：

- 少量跨用例稳定共享、且不带基础设施倾向的类型

边界约束：

- 默认不新增共享包
- 只有当类型被多个用例稳定共享，且放回单一用例会造成反向依赖时，才允许进入此层

## 现有能力的迁移落位

### `auth`

- 用例迁入 `server/service/system/usecase/auth.go`
- `AuthStore` 接口仍由 `auth` 用例定义
- `GormAuthStore` 迁入 `server/service/system/infra/persistence/auth_store_gorm.go`
- JWT 实现迁入 `server/service/system/infra/identity/jwt.go`
- LDAP 客户端实现迁入 `server/service/system/infra/directory/ldap_client.go`
- 当前 `server/service/system/jwt.go` 在本轮保留为薄兼容层，只做转发，不继续承载真实实现

### JWT 兼容策略

本轮明确采用：

- `JWT` 真实实现迁入 `server/service/system/infra/identity/jwt.go`
- `server/service/system/jwt.go` 暂时保留为薄包装，向旧调用方暴露兼容函数
- `server/middleware/jwt.go` 与现有相关测试本轮允许继续走兼容层，不在本轮引入 `TokenParser` capability

本轮明确不做：

- 不让 `middleware` 直接依赖 `infra/identity`
- 不把 JWT middleware 的能力边界重构与 `问题 2` 绑定在一起

后续处理方向：

- 等 `system` 域内部边界稳定后，再在后续问题中评估是否把 token 解析进一步收敛为 capability，并与认证/授权治理一起处理

### `user`

- `UserService` 迁入 `server/service/system/usecase/user.go`
- `GormUserStore` 迁入 `server/service/system/infra/persistence/user_store_gorm.go`

### `role`

- `RoleService` 迁入 `server/service/system/usecase/role.go`
- `GormRoleStore` 迁入 `server/service/system/infra/persistence/role_store_gorm.go`

### `ldap`

- `LDAPService` 迁入 `server/service/system/usecase/ldap.go`
- `GormLDAPStore` 迁入 `server/service/system/infra/persistence/ldap_store_gorm.go`
- LDAP 网络调用实现复用 `server/service/system/infra/directory/ldap_client.go`

### `audit`

- `AuditService` 迁入 `server/service/system/usecase/audit.go`
- `GormAuditStore` 迁入 `server/service/system/infra/persistence/audit_store_gorm.go`

### `rbac`

- `RBACService` 迁入 `server/service/system/usecase/rbac.go`
- `GormRBACStore` 迁入 `server/service/system/infra/persistence/rbac_store_gorm.go`

### `server/service/system` 根包兼容策略

本轮明确采用“直接迁移调用方 + 极小兼容层”：

- `server/api/v1/system`、`server/initialize`、`server/middleware` 以及相关测试在本轮内直接迁移到新的 `usecase` / `infra` 子包依赖
- `server/service/system` 根包不作为长期 façade 保留
- 根包仅允许保留以下两类内容：
  - 本轮暂不纳入主体的辅助能力，如 `seed.go`、`health.go`、`password.go`
  - 为压缩迁移爆炸面而临时保留的极小兼容包装，例如 `jwt.go`

根包明确不再承载：

- 新的业务用例实现
- 新的 Gorm store 实现
- 新的 LDAP / JWT 真实适配器实现
- 面向外部调用方的长期过渡 façade

### 暂不纳入本轮主体

以下文件不作为 `问题 2` 的主改造对象，只在需要维持编译或引用路径时做最小配合：

- `seed.go`
- `health.go`
- `password.go`

原因：

- 它们不是“用例与基础设施混包”的主要矛盾
- 本轮优先聚焦在能力边界，而不是顺手重写所有系统辅助能力

## 错误边界

本轮遵循“结构变，行为尽量不变”的原则。

- `usecase` 返回领域错误或稳定语义错误
  - 例如 `ErrInvalidCredentials`、`ErrUserDisabled`、`ErrProtectedAdminRole`
- `infra` 返回实现相关错误
  - 数据库、LDAP、JWT 相关错误先按实现语义上抛
- `api` 继续负责将错误映射为 HTTP 状态码与响应体
- `middleware` 只处理上下文桥接、鉴权结果和自身错误路径，不新增业务规则判断

关键约束：

- 不把 HTTP 语义带进 `usecase`
- 不把业务规则判断塞进 `infra`

## 测试策略

本轮验收不以“目录拆完”为标准，而以“行为不退化、边界变清晰”为标准。

### 1. 用例测试

保留并强化现有 `fake store / fake dependency` 形状：

- `AuthService`
- `UserService`
- `RoleService`
- `LDAPService`
- `RBACService`
- `AuditService`

目标：

- 证明业务规则、审计触发、错误语义未变
- 避免用例测试依赖 Gorm

### 2. 基础设施测试

为被拆出的适配器建立独立测试：

- `Gorm*Store` 只测查询、事务、排序、去重、记录不存在等持久化语义
- JWT 适配器单独测试签发/解析行为
- LDAP 客户端单独测试连接和认证适配语义

### 3. 组合边界测试

继续保留并补强：

- `initialize/system_module_test.go`
- `router` 注入相关测试
- `middleware` 注入相关测试

目标：

- 证明 `SystemModule` 仍然是唯一装配边界
- 证明 public/private 路由与授权链行为未变

## 改造顺序

建议按以下顺序执行，避免一次性爆炸式迁移：

1. 先建立 `server/service/system/usecase` 与 `server/service/system/infra/*` 目录
2. 迁移 `auth / user / role / ldap / audit / rbac` 的接口与用例实现
3. 迁移对应 `Gorm*Store`、JWT、LDAP 客户端实现
4. 调整 `initialize.SystemModule`，改为装配新边界
5. 调整 `api` / `middleware` / 其他调用点的 import 与注入
6. 删除或收薄旧 `server/service/system` 中的兼容残留
7. 运行测试并验证对外行为不变

## 风险

### 风险 1：共享类型包退化成新杂物间

规避方式：

- 默认不创建共享类型
- 只有确实被多个用例稳定共享时才上移

### 风险 2：用例层仍然偷偷依赖实现细节

规避方式：

- 在迁移时逐个检查 import
- `usecase` 包内禁止出现 `gorm`、LDAP 客户端实现、JWT 库细节

### 风险 3：问题 2 与问题 3、5 再次纠缠

规避方式：

- 本轮只修 `system` 域的内部边界
- 不顺手消灭 `global`
- 不顺手重构为全项目级 capability 系统

### 风险 4：结构变了，但行为测试没锁住

规避方式：

- 先保留现有用例测试形状
- 再为拆出的适配器补独立测试
- 最后通过 `SystemModule` 和 router/middleware 测试锁组合边界

## 结果预期

完成本设计后，系统不会立刻变成完整的平台内核，但会得到一个明确的结构收益：

- `system` 域内部不再把“系统做什么”和“系统靠什么实现”混在同一个包边界里
- 后续替换持久化、LDAP、JWT 实现时，改动面将主要落在 `infra`
- 后续继续推进 `global` 收缩、治理能力独立化时，会有更清晰的能力落点

一句话概括：

`把问题 2 从“类型上看起来分层”推进到“包边界上真的分层”。`
