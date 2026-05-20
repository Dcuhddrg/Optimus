# System Module Composition Design

日期：2026-05-11

## 背景

在当前 `server/` 后端中，`api` 和部分 `middleware` 同时承担了协议适配与依赖装配职责。典型表现包括：

- handler 内判断 `global.OPTIMUS_DB` 是否可用
- handler 内现场构造 `NewGormXxxStore(...)`
- handler 内现场构造 `NewXxxService(...)`
- 鉴权中间件内现场构造授权服务

这导致装配逻辑按 endpoint 和请求链扩散，边界不稳定，不利于未来继续演进为可治理的平台内核。

本设计只处理“问题 1：依赖装配落在 API / 中间件层”。

本设计不直接解决以下问题：

- `service` 混合用例逻辑与基础设施实现
- `global` 作为隐式依赖总线
- 启动流程混合运行时装配与业务副作用
- 认证、授权、审计能力的进一步独立治理

这些问题在本设计中只做兼容性约束，不在本轮一起重构。

## 目标

本设计的目标是把 `system` 域相关依赖的装配权从 `api` / `middleware` 收回到系统边缘，同时保持现有 `initialize / router / api / service` 外层工程结构可延续。

成功标准：

- `api` 层不再现场构造 store / service / client
- `middleware` 不再通过 `global` 现场构造业务授权服务
- `router` 只做路径挂接，不参与依赖拼装
- `system` 域形成一个集中装配入口，供路由和中间件消费
- 本轮改造不要求引入新的总体分层，如完整 `application/infra` 目录重构

## 非目标

- 不引入全项目级超级容器
- 不把所有业务模块统一重构为新模块系统
- 不在本轮中拆分 `service/system` 中的 Gorm 实现
- 不在本轮中删除 `global`
- 不在本轮中调整接口协议、路由路径、权限语义

## 方案选择

本轮采用“中等重构”方案：`SystemModule` 统一装配。

不采用更保守方案的原因：

- 仅把 `new store/new service` 从 handler 挪到别处，容易形成新的分散手工装配点，不能真正收回装配权

不采用更彻底方案的原因：

- 直接上完整 composition root + transport adapter 会把问题 1 与问题 2、3 混在一起，当前收益与改动成本不匹配

## 设计概览

新增一个只服务于 `system` 域的模块装配对象，暂定名 `SystemModule`。

它的唯一职责是：

`作为 system 域的运行时装配边界，把已经组装好的能力统一暴露给 router、api、middleware 使用。`

新的责任分布如下：

- `initialize`
  - 负责创建 `SystemModule`
- `router`
  - 负责挂接 `SystemModule` 提供的 handlers 和 middleware
- `api`
  - 只做协议适配：收参、调用用例、映射错误、写回响应
- `middleware`
  - 只做上下文桥接和调用已注入 capability
- `service`
  - 继续承载现有用例逻辑与实现，作为本轮兼容层

## `SystemModule` 的职责边界

### 应该承担的职责

1. 统一构造 `system` 域依赖
2. 暴露稳定消费面给 router / api / middleware
3. 成为后续继续收紧边界的落脚点

### 不应该承担的职责

1. 不成为新的全局变量
2. 不承载业务逻辑
3. 不吞掉整个项目的装配职责
4. 不把内部装配材料暴露给外部继续拼装

## `SystemModule` 的对外暴露面

`SystemModule` 只暴露三类对象。

### 1. Handlers

给 router 使用，代表已经准备好的 HTTP 处理器。

建议按资源或能力拆分，而不是继续维持单个大 `Api` 空壳：

- `AuthHandler`
- `UserHandler`
- `RoleHandler`
- `LDAPHandler`
- `AuditHandler`
- `BaseHandler`
- `HealthHandler`

router 只做路径与 handler 方法绑定。

### 2. Capabilities

给 middleware 或未来其他入口使用，只暴露稳定能力接口，不暴露装配材料。

当前明确需要的能力：

- `RBACAuthorizer`

后续可能扩展但本轮不必一并实现：

- `TokenParser`
- `CurrentIdentityResolver`

### 3. Services

作为 handler 的内部依赖层存在，但不要求 router 或外部直接消费全部 service。

原则：

- handler 可以持有 service
- module 不应把 `*gorm.DB`、原始 store、原始 client、原始配置对象直接暴露出去

## handler 设计

handler 从“请求时装配依赖”改为“启动时接收依赖”。

### 设计原则

1. handler 是带依赖的对象
2. handler 按能力拆分，不坚持单一 `Api` 聚合
3. 错误映射仍留在 handler
4. handler 不再碰 `global`

### handler 应保留的职责

- 解析请求
- 调用 service / capability
- 将领域或用例错误映射为 HTTP 状态码与响应体

### handler 应移除的职责

- 判断数据库或底层依赖是否存在
- 现场 `new store/new service/new client`
- 决定依赖组合方式

## router 设计

router 只接收一个已经准备好的 `SystemModule`，并把路径挂接到模块提供的 handlers。

新的 router 边界：

- 定义 public/private 路由分组
- 绑定路径到 handler 方法
- 组合由外部注入的 middleware

router 不再承担：

- new handler 空壳
- 推断 handler 依赖来源
- 参与业务依赖装配

## middleware 设计

middleware 不再在请求处理中现场构造业务服务。

### JWT middleware

本轮不要求彻底抽成完整认证 capability，但要遵守同一方向：

- 中间件应尽量消费外部注入的能力
- 不继续扩散新的业务装配逻辑

### RBAC middleware

当前是本轮重点。

新的边界应该是：

- 中间件从外部接收已装配的 `RBACAuthorizer`
- 中间件仅负责从上下文中取 claims、调用授权能力、按结果决定放行或拒绝

中间件不再承担：

- 读取 `global.OPTIMUS_DB`
- 现场构造 `RBACService`
- 决定授权能力的组合方式

## 依赖流向

本轮目标依赖流向为：

`initialize -> build SystemModule -> router mounts handlers/middleware -> api/middleware consume injected dependencies -> service executes use cases`

关键约束：

- 请求处理链上不再产生新的业务依赖装配动作
- 装配动作集中在系统边缘完成

## 改造顺序

建议顺序如下：

1. 定义 `SystemModule` 结构与构造入口
2. 定义 `Handlers` / `Capabilities` 暴露面
3. 将 `system` 下各 handler 改为显式持有依赖的对象
4. 调整 `router/system`，改为消费 `SystemModule`
5. 调整 `RBAC` middleware，改为消费注入 capability
6. 清理 `api` / `middleware` 中残留的现场装配逻辑
7. 验证现有 HTTP 接口行为不变

## 兼容性约束

本轮需要保持以下内容不变：

- 现有路由路径
- 现有 public/private 分组语义
- 现有 `service` 逻辑与错误语义
- 现有测试目标与已有对外行为

## 风险

### 风险 1：模块对象退化成新“大泥球”

规避方式：

- 只暴露 handlers 与少量 capabilities
- 不把原始依赖材料对外暴露

### 风险 2：把问题 1 和问题 2 一起改炸

规避方式：

- 本轮不调整 `service/system` 文件分层
- 把 `service` 视为兼容层，先只收装配位置

### 风险 3：中间件与 handler 改造风格不一致

规避方式：

- 明确“请求处理链上不再做装配”作为统一规则

## 结果预期

完成本设计后，系统不会立刻成为严格的平台内核，但会得到一个重要的结构收益：

- 装配权回到系统边缘
- `api` 回到协议适配层
- `middleware` 回到桥接层
- 后续继续处理问题 2、3、4、5 时，有稳定落脚点而不是到处改请求入口

这一步的意义不是完成最终架构，而是先把“谁有权拼装能力”这件事定下来。
