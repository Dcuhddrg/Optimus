# 后端分层与迁移规则

> 项目规则以 [AGENT.MD](../../AGENT.MD) 为准。

## 后端分层契约

新模块与存量模块统一遵循以下分层、目录和依赖方向：

```text
Router -> API -> Service -> Model
```

| 层 | 职责与位置 | 依赖边界 |
| --- | --- | --- |
| Router | `server/router/<module>/`，注册路径、中间件与 public/private 分组 | 连接 API，不承担业务处理 |
| API | `server/api/v1/<module>/`，绑定和校验 HTTP 输入、调用服务、组织响应 | 通过 Service 处理业务，不绕过服务直接访问持久化 |
| Service | `server/service/<module>/`，业务规则、流程编排和数据访问协作 | 使用 Model 及所需基础设施，不依赖 `gin.Context`，不反向调用 API/Router |
| Model | `server/model/<module>/`，实体、请求/响应数据结构和持久化映射 | 不反向依赖 Router/API/Service |

- 保留现有 `enter.go` 分组注册和组合入口，新增分组沿用对应模块的注册方式；不得绕过已有入口另建平行注册链。
- 模块装配按下文的 GroupApp 目标入口统一；切换时同批删除被替代的 handlers/Module 平行装配链，保留必要能力。
- API 向 Service 显式传入所需业务参数。Swagger 注释与真实接口行为保持一致。
- 新旧模块均按上述四层职责组织；原始参考模板见 [开发示例](../examples/README.md)，迁移取舍和验证在 [#45](https://github.com/Dcuhddrg/Optimus/issues/45) 跟进；#8 为历史示例交付。

### 后端结构维护要求

- 不得重新引入已清理的旧 `usecase/infra` 分层；新发现的结构差异在 GitHub Issue 跟进。
- 数据库查询和事务归对应职责 Service，生产路径使用 `global.OPTIMUS_DB` 或当前事务句柄。需要替换的测试依赖与外部客户端可以使用接口，启动装配不创建平行 Store 实现。
- 外部 SDK、协议适配和公共能力按职责迁入明确的模块或插件位置，每项给出旧路径、新路径、调用者与测试映射；避免仅给旧层改名后保留平行体系。
- 同批调整 API 导入、服务构造、`enter.go`、初始化注入和测试。必要接口可保留用于解耦与测试，不能成为旧包长期存在的豁免理由。
- 结构调整保持已有接口返回、HTTP 状态、权限和项目隔离行为；响应协议改造由 #18 独立实施。
- 结构调整须清除临时兼容入口，并按影响范围验证构建和隔离 MySQL 回归；进度与剩余差异在 GitHub 跟进。

现行决定及替代关系见 [ADR-002](../memory/long-term/decisions/ADR-002-backend-structure-migration.md)，迁移结果及验证依据见 [结构迁移 #11](https://github.com/Dcuhddrg/Optimus/issues/11)。

## 装配与注册入口

用户已确认后端迁移参考 gin-vue-admin。目标为 `service/enter.go` 的 ServiceGroupApp、`api/v1/enter.go` 的 ApiGroupApp 与 `router/enter.go` 的 RouterGroupApp 根级聚合；各模块 `enter.go` 聚合职责 Service/API/Router。API 通过模块统一别名访问 ServiceGroup，Router 通过统一别名访问 ApiGroup，具体路径绑定放在职责文件。请求 DTO 进入 `model/<module>/request/`，公共请求/响应进入 `model/common/request/` 与 `model/common/response/`。详细生命周期、兼容边界与批次见[迁移设计](../superpowers/specs/2026-11-02-gva-backend-migration-design.md)。

聚合与职责声明须分开：enter.go 中仅定义 ApiGroup、ServiceGroup 或 RouterGroup；AuthApi/UserApi 等职责类型位于 auth.go/user.go 等对应文件，Authorizer、StatusService 等支持接口也归对应职责文件。行为测试通过不等于组织方式对齐，验收同时检查声明位置与聚合边界。

HTTP 输入包括完整表单、局部状态/角色/权限操作和查询条件，均使用模块 request 包中的命名类型。审计、资产等分页查询 DTO 内嵌公共 PageInfo；API 负责按既有协议解析并映射到内部 Service 过滤器。Service 条件可保留项目范围等业务字段，但不能用它代替 HTTP request DTO，也不接受客户端伪造的已授权项目列表。

`main` 与 `initialize` 仍负责启动、配置和基础设施准备，再装配唯一的 ServiceGroup，最后注册路由。依赖只在启动期初始化，GroupApp 的发布对象和集中别名保持稳定；请求处理不得重新装配全局依赖，测试须隔离并恢复共享状态。数据库与外部客户端按生命周期由 initialize 传入或从现有 global 入口取得，构造函数可以用于组内部依赖创建，不再作为 Router 逐次注入 handlers 的并行生产入口。

当前生产装配已切换为 `initialize.InitializeServices → ServiceGroupApp → ApiGroupApp → RouterGroupApp`；旧 `New*Module/New*Handlers` 入口已删除，审计清理任务与 HTTP API 使用同一 AuditService。现有 GroupApp 是 #45 按上游组织方式落地的新入口，不恢复 #11 已清理的旧 usecase/infra。模块内使用构造函数装配 LDAP、云厂商客户端等外部依赖。数据库查询和事务使用启动时发布的共享句柄，现有替身接口服务于测试。HTTP 状态、字段、分页、表结构和删除行为维持已确认契约；插件注册框架另行设计。

### Router 组织约定

所有后端模块统一采用按职责拆分子路由、由模块 `enter.go` 聚合的组织方式，适用于 system、project、cloud、插件及后续新增模块，新增与存量代码均须遵守。参考 gin-vue-admin 的 [system/enter.go](https://github.com/flipped-aurora/gin-vue-admin/blob/main/server/router/system/enter.go)，其中 system 仅为示例，不限定规范适用范围：

- `router/enter.go` 汇总 system、project、cloud 等模块的 `RouterGroup`。
- `router/<module>/enter.go` 通过匿名嵌入聚合子 Router，并集中声明 API 别名。`initialize/router.go` 直接调用各职责 `InitXxxRouter`，按原顺序传入公开、私有或项目范围组。
- 具体路径、中间件和 handler 绑定放在同目录职责文件中，按各模块实际职责命名，不照搬 system 的业务划分。例如 system 拆为 `base.go`、`auth.go`、`user.go`、`role.go`、`ldap.go`、`audit.go`，角色权限及权限树由 `RoleRouter` 注册；project 可按项目与成员、cloud 可按账号与资产组织子 Router。
- 子 Router 参考上游从模块 enter.go 的集中别名取得 ApiGroupApp 中对应 API；API 从 ServiceGroupApp 的模块聚合取得服务，不再要求生产路由逐次显式注入 handler。初始化顺序、别名稳定性及测试替换按迁移设计执行；不允许两套生产装配链长期并存。
- 公开与私有组由 initialize 创建并传入，保持路径、HTTP 方法、注册顺序、JWT/RBAC、项目范围及响应协议标记的实际作用范围。拆分不能重复挂载中间件、重复注册或放宽权限。

本约定适用于全仓库。system、project、cloud 的模块 enter.go 聚合子 Router 类型与 API 别名，initialize 调用职责注册方法。实现批次、验证边界和待验收事项见 [#45](https://github.com/Dcuhddrg/Optimus/issues/45)。结构整改与 #18 响应协议迁移分批提交。原始参考示例见 [Router](../examples/backend/router-example.md) 与 [enter.go](../examples/backend/enter-go-example.md)。
