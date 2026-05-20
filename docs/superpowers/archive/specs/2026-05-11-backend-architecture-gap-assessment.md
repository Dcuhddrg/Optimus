# Optimus 后端架构问题清单

日期：2026-05-11

范围：
- 仅评估 `server/` 后端
- 参考对象为 `gin-vue-admin`
- 评估标准高于 `gin-vue-admin` 的后台脚手架目标，偏向未来可内核化的平台服务

结论：

Optimus 当前后端的主要问题，不是分层名称错误，而是处在一个中间态：

- 外层采用了 `initialize / router / api / service / global` 这套后台框架常见组织
- 工程配套和边界稳定性还没有达到成熟后台基座的程度
- 同时又没有进一步收紧成适合平台内核演进的结构

因此它更像“轻量骨架 + 局部手工拼装”，而不是“稳定后台基座”或“可治理的平台内核”。

## 严重度排序

1. 依赖装配落在 API / 中间件层
2. `service` 混合用例逻辑与基础设施实现
3. `global` 作为隐式依赖总线
4. 启动流程混合运行时装配与业务副作用
5. 权限 / 认证能力还挂在 `system` 业务模块下

## 问题 1：依赖装配落在 API / 中间件层

这是首要矛盾。它决定了系统后续会沿什么模式继续复制。

当前表现：

- `handler` 自己判断 `global.OPTIMUS_DB` 是否存在
- `handler` 现场构造 `NewGormXxxStore(...)`
- `handler` 现场构造 `NewXxxService(...)`
- 中间件内部也会现场构造业务授权服务
- 项目里同时存在“聚合入口调用”和“请求入口现场装配”两套模式

代表位置：

- `server/api/v1/system/auth.go`
- `server/api/v1/system/user.go`
- `server/api/v1/system/role.go`
- `server/api/v1/system/ldap.go`
- `server/api/v1/system/audit.go`
- `server/middleware/casbin_rbac.go`

为什么会阻碍未来内核化：

- HTTP 层会逐渐变成事实上的控制层
- 模块装配逻辑按 endpoint 扩散，无法按模块收敛
- 新增 gRPC、任务、CLI 等入口时，能力无法自然复用
- 统一治理会越来越依赖“每个入口都记得照做”

一句话概括：

`请求入口正在夺走系统边缘本该拥有的装配权。`

## 问题 2：`service` 混合用例逻辑与基础设施实现

这是第二优先级问题。短期不一定出故障，但长期一定卡住演进。

当前表现：

- `service/system` 中既放用例逻辑，也放 Gorm store 实现
- 稳定业务规则和持久化细节位于同一包、同一文件中
- 接口虽然存在，但包边界并没有真正隔离能力与实现

代表位置：

- `server/service/system/user.go`
- `server/service/system/auth.go`
- `server/service/system/role.go`
- `server/service/system/ldap.go`

为什么会阻碍未来内核化：

- 难以区分“稳定能力”与“可替换实现”
- 业务规则容易被 ORM 和表结构反向塑形
- 引入缓存、异步化、读写拆分、存储替换时改动面过大
- 测试会越来越依赖整块数据库上下文，而不是按能力边界测试

一句话概括：

`抽象已经写在类型名上，但还没有落实在包边界上。`

## 问题 3：`global` 作为隐式依赖总线

这是常见脚手架做法，但对平台内核目标伤害较大。

当前表现：

- 配置、日志、数据库、Redis、Casbin 等挂在 `global` 下
- 模块可以通过全局状态直接获取运行时依赖
- 显式依赖关系在调用侧被弱化

代表位置：

- `server/global/global.go`
- `server/api/v1/system/*.go`
- `server/middleware/*.go`
- `server/initialize/*.go`

为什么会阻碍未来内核化：

- 模块边界会变成“谁都能摸到共享状态”
- 不利于裁剪能力、隔离运行模式和多实例测试
- 表面模块化，实际耦合在进程级运行时上

一句话概括：

`全局单例让系统更像共享进程，而不是可治理的能力组合。`

## 问题 4：启动流程混合运行时装配与业务副作用

这个问题平时不会最先暴露，但在部署拓扑变复杂时会迅速放大。

当前表现：

- 启动阶段既做 DB / Redis / Casbin 初始化
- 又做 `AutoMigrateSystem`
- 又做 `SeedSystemData`

代表位置：

- `server/initialize/startup.go`

为什么会阻碍未来内核化：

- 运行时装配、基础设施探活、业务初始化、迁移任务没有分开
- 未来引入只读节点、任务节点、灰度环境、多租户时边界会混乱
- 进程角色难以清晰区分

一句话概括：

`进程启动和业务初始化任务耦合在了一起。`

## 问题 5：权限 / 认证能力还挂在 `system` 业务模块下

这是一个重要问题，但更多是前面几个问题共同作用的结果。

当前表现：

- JWT 解析直接依赖 `systemService.ParseJWT(...)`
- RBAC 鉴权逻辑通过中间件临时组装 `RBACService`
- 认证、授权、审计还没有形成独立 runtime capability

代表位置：

- `server/middleware/jwt.go`
- `server/middleware/casbin_rbac.go`
- `server/service/system/jwt.go`
- `server/service/system/rbac.go`
- `server/service/system/audit.go`

为什么会阻碍未来内核化：

- 认证、授权、审计能力难以统一治理
- 非 HTTP 入口难以复用同一套能力边界
- 权限能力和 `system` 业务域绑定过紧

一句话概括：

`治理能力还没有从具体业务模块里站出来。`

## 目标边界轮廓

不进入具体改造方案，仅记录目标方向：

- `runtime composition`
  - 负责装配依赖、注册模块、决定进程角色
- `transport adapters`
  - HTTP 只做协议转换和上下文桥接
- `application use cases`
  - 用户、角色、认证、LDAP、审计等能力的用例编排
- `domain policies`
  - admin 保护、状态约束、权限语义等稳定规则
- `infrastructure adapters`
  - Gorm/MySQL、JWT、LDAP、Casbin、Redis 等实现细节

## 后续处理建议

建议按严重度从上往下逐项解决，不要并行大改：

1. 先收敛依赖装配位置
2. 再拆开用例层与基础设施实现
3. 再减少 `global` 在业务路径中的存在感
4. 再清理启动流程中的副作用职责
5. 最后再把认证、授权、审计沉淀成独立治理能力
