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

- `enter.go` 聚合分组，`initialize` 注册职责路由；不得另建平行装配或注册链。
- API 向 Service 显式传入所需业务参数。Swagger 注释与真实接口行为保持一致。
- 新旧模块均按上述四层职责组织；原始参考模板见 [开发示例](../examples/README.md)，迁移取舍和验证在 [#45](https://github.com/Dcuhddrg/Optimus/issues/45) 跟进；#8 为历史示例交付。

### 后端结构维护要求

- 不得重新引入已清理的旧 `usecase/infra` 分层；新发现的结构差异在 GitHub Issue 跟进。
- 数据库查询和事务归对应职责 Service，生产路径使用 `global.OPTIMUS_DB` 或当前事务句柄。需要替换的测试依赖与外部客户端可以使用接口，启动装配不创建平行 Store 实现。
- 外部 SDK、协议适配和公共能力按职责迁入明确的模块或插件位置，每项给出旧路径、新路径、调用者与测试映射；避免仅给旧层改名后保留平行体系。
- 同批调整 API 导入、服务构造、`enter.go`、初始化注入和测试。必要接口可保留用于解耦与测试，不能成为旧包长期存在的豁免理由。
- 结构调整保持已有接口返回、权限和项目隔离行为；响应协议照 GVA，见[接口契约](../frontend-backend/boundary.md#4-响应协议与-http-语义)。
- 结构调整须清除临时兼容入口，并按影响范围验证构建和隔离 MySQL 回归；进度与剩余差异在 GitHub 跟进。

现行决定及替代关系见 [ADR-002](../memory/long-term/decisions/ADR-002-backend-structure-migration.md)，迁移结果及验证依据见 [结构迁移 #11](https://github.com/Dcuhddrg/Optimus/issues/11)。

## 装配与注册入口

根级通过 ServiceGroupApp、ApiGroupApp 和 RouterGroupApp 汇总模块；模块 `enter.go` 聚合职责类型和集中别名。API 通过别名访问 Service，Router 通过别名访问 API。详细设计与兼容边界见[迁移设计](../superpowers/specs/2026-11-02-gva-backend-migration-design.md)。

聚合与职责声明须分开：enter.go 中仅定义 ApiGroup、ServiceGroup 或 RouterGroup；AuthApi/UserApi 等职责类型位于 auth.go/user.go 等对应文件，Authorizer 等支持接口也归对应职责文件。ServiceGroup 匿名嵌入具体服务值，API 的 Service 别名与 Router 的 API 别名复制聚合值。行为测试通过不等于组织方式对齐，验收同时检查声明位置与聚合边界。

HTTP 输入包括完整表单、局部状态/角色/权限操作和查询条件，均使用模块 request 包中的命名类型。审计、资产等分页查询 DTO 内嵌公共 PageInfo；API 负责按既有协议解析并映射到内部 Service 过滤器。Service 条件可保留项目范围等业务字段，但不能用它代替 HTTP request DTO，也不接受客户端伪造的已授权项目列表。

按 GVA，Service 是无状态值：结构体不含字段，不提供 `NewXxxService` 构造函数，也不在启动期装配。数据库、配置和外部客户端在调用时从 `global` 读取，或像 GVA 的 `utils.NewJWT()`、`upload.NewOss()` 一样按配置即时构建（如 `utils/ldap.NewLDAP()`）；服务间调用使用 `XxxServiceApp` 包级变量。API 别名在包初始化时复制 ServiceGroupApp 中的值，不再刷新。`main` 与 `initialize` 只负责配置、日志、数据库、数据初始化及 provider 注册（如 `initialize.Cloud()`），随后由 `core.RunServer()` 注册路由。测试替换依赖时修改 `global` 或包内测试钩子并在 Cleanup 中恢复；`initialize` 的契约测试会拒绝带字段的 Service。

当前链路为 `ServiceGroupApp → ApiGroupApp → RouterGroupApp`，见 [#52](https://github.com/Dcuhddrg/Optimus/issues/52)。审计清理任务和 HTTP API 使用同一审计职责及数据库；插件注册框架另行设计。迁移与验证记录见 [#45](https://github.com/Dcuhddrg/Optimus/issues/45) 和 [#48](https://github.com/Dcuhddrg/Optimus/pull/48)。

### Router 组织约定

所有后端模块按职责拆分子路由，参考 gin-vue-admin 的 [system/enter.go](https://github.com/flipped-aurora/gin-vue-admin/blob/main/server/router/system/enter.go)：

- `router/enter.go` 汇总 system、project、cloud 等模块的 `RouterGroup`。
- `router/<module>/enter.go` 通过匿名嵌入聚合子 Router，并集中声明 API 别名。`initialize/router.go` 直接调用各职责 `InitXxxRouter`，按原顺序传入公开、私有或项目范围组。
- 具体路径、中间件和 handler 绑定放在同目录职责文件中，按各模块实际职责命名，不照搬 system 的业务划分。例如 system 拆为 `base.go`、`auth.go`、`user.go`、`role.go`、`ldap.go`、`audit.go`，角色权限及权限树由 `RoleRouter` 注册；project 可按项目与成员、cloud 可按账号与资产组织子 Router。
- 公开与私有组由 initialize 创建并传入，保持路径、HTTP 方法、注册顺序、JWT/RBAC、项目范围及响应协议标记的实际作用范围。拆分不能重复挂载中间件、重复注册或放宽权限。

原始参考示例见 [Router](../examples/backend/router-example.md) 与 [enter.go](../examples/backend/enter-go-example.md)；结构整改与 #18 响应协议迁移分批提交。
