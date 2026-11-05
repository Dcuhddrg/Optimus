# 后端分层规则

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

- `enter.go` 作为分组注册与组合入口，聚合职责类型和别名；具体路由由 `initialize` 注册，不另建平行装配或注册链。
- API 向 Service 显式传入所需业务参数。Swagger 注释与真实接口行为保持一致。
- 各层写法示例见 [开发示例](../examples/README.md)。

### 结构调整要求

- 不得重新引入已清理的旧 `usecase/infra` 分层；新发现的结构差异登记到 GitHub Issue。
- 数据库查询和事务归对应职责 Service，生产路径使用 `global.OPTIMUS_DB` 或当前事务句柄。需要替换的测试依赖与外部客户端可以使用接口，启动装配不创建平行 Store 实现。
- 外部 SDK、协议适配和公共能力按职责迁入明确的模块或插件位置，每项给出旧路径、新路径、调用者与测试映射；避免仅给旧层改名后保留平行体系。
- 同批调整 API 导入、服务构造、`enter.go`、初始化注入和测试。必要接口可保留用于解耦与测试，不能成为旧包长期存在的豁免理由。
- 结构调整保持已有接口返回、权限和项目隔离行为；响应协议照 GVA，见[接口契约](../frontend-backend/boundary.md#4-响应协议与-http-语义)。
- 结构调整须清除临时兼容入口，并按影响范围验证构建和隔离 MySQL 回归。

决定与替代关系见 [架构决策索引](../adr/README.md)。

## 装配与注册入口

根级通过 ServiceGroupApp、ApiGroupApp 和 RouterGroupApp 汇总模块，链路为 `ServiceGroupApp → ApiGroupApp → RouterGroupApp`；模块 `enter.go` 聚合职责类型和集中别名。API 通过别名访问 Service，Router 通过别名访问 API。

聚合与职责声明须分开：enter.go 中仅定义 ApiGroup、ServiceGroup 或 RouterGroup；AuthApi/UserApi 等职责类型位于 auth.go/user.go 等对应文件，Authorizer 等支持接口也归对应职责文件。ServiceGroup 匿名嵌入具体服务值，API 的 Service 别名与 Router 的 API 别名复制聚合值。行为测试通过不等于组织方式对齐，验收同时检查声明位置与聚合边界。

HTTP 输入包括完整表单、局部状态/角色/权限操作和查询条件，均使用模块 request 包中的命名类型。审计、资产等分页查询 DTO 内嵌公共 PageInfo；API 负责按既有协议解析并映射到内部 Service 过滤器。Service 条件可保留项目范围等业务字段，但不能用它代替 HTTP request DTO，也不接受客户端伪造的已授权项目列表。

按 GVA，Service 是无状态值：结构体不含字段，不提供 `NewXxxService` 构造函数，也不在启动期装配。数据库、配置和外部客户端在调用时从 `global` 读取，或像 GVA 的 `utils.NewJWT()`、`upload.NewOss()` 一样按配置即时构建（如 `utils/ldap.NewLDAP()`）；服务间调用使用 `XxxServiceApp` 包级变量。API 别名在包初始化时复制 ServiceGroupApp 中的值，不再刷新。`main` 与 `initialize` 只负责配置、日志、数据库、数据初始化及 provider 注册（如 `initialize.Cloud()`），随后由 `core.RunServer()` 注册路由。测试替换依赖时修改 `global` 或包内测试钩子并在 Cleanup 中恢复；`initialize` 的契约测试会拒绝带字段的 Service。

种子数据按 GVA 放在 `source/<模块>`，用 `RegisterInit(order, initializer)` 注册，由 `initialize.InitializeData` 在启动时按序迁移并幂等写入；定时任务实现放在 `task/`，在 `initialize.Timer` 用 `global.OPTIMUS_Timer` 注册；GORM 配置与日志放在 `initialize/internal`。

插件的注册与装配见[插件开发约束](plugin-development.md)。

### Router 组织约定

所有后端模块按职责拆分子路由，参考 gin-vue-admin 的 [system/enter.go](https://github.com/flipped-aurora/gin-vue-admin/blob/main/server/router/system/enter.go)：

- `router/enter.go` 汇总 system、project、cloud 等模块的 `RouterGroup`。
- `router/<module>/enter.go` 通过匿名嵌入聚合子 Router，并集中声明 API 别名。`initialize/router.go` 直接调用各职责 `InitXxxRouter`，按原顺序传入公开、私有或项目范围组。
- 具体路径、中间件和 handler 绑定放在同目录职责文件中，按各模块实际职责命名，不照搬 system 的业务划分。例如 system 拆为 `base.go`、`auth.go`、`user.go`、`role.go`、`ldap.go`、`audit.go`，角色权限及权限树由 `RoleRouter` 注册；project 可按项目与成员、cloud 可按账号与资产组织子 Router。
- 公开与私有组由 initialize 创建并传入，保持路径、HTTP 方法、注册顺序、JWT/RBAC、项目范围及响应协议标记的实际作用范围。拆分不能重复挂载中间件、重复注册或放宽权限。

写法示例见 [Router](../examples/backend/router-example.md) 与 [enter.go](../examples/backend/enter-go-example.md)。

## 各层通用写法

参照 gin-vue-admin `cee04b08` 的 `aiDoc/modules/backend-layer-rules.md`，与上文的装配约定一起使用。

### Model 层

- 实体优先嵌入 `global.OPTIMUS_MODEL`（对应 GVA 的 `GVA_MODEL`），`ID`、`CreatedAt`、`UpdatedAt` 沿用其约定；关联表等确需自定义主键的，保持字段名和 JSON 名一致。
- 字段补全 `gorm` 与 `json` 标签，JSON 名用 camelCase，敏感字段用 `json:"-"`；表名通过 `TableName()` 显式声明。
- 请求 DTO 放在 `model/<module>/request/`，响应 DTO 放在 `model/<module>/response/`；列表查询定义 `XxxSearch` 并内嵌 `model/common/request.PageInfo`。

### 类型一致性

- 同一字段在实体、请求 DTO、响应 DTO 和前端类型中保持一致；ID、状态、枚举和时间字段重点核对。
- 指针与非指针互转（如可空时间 `*time.Time`）在 Service 层显式处理 `nil`，不让零值静默写入。

### API 参数来源

绑定方式与参数真实来源一致，不为了套模板把 Header、Cookie、Query 或 form-data 的数据改成 body。

| 来源 | 取法 |
| --- | --- |
| JSON body | `ShouldBindJSON` |
| Query | `ShouldBindQuery`、`c.Query`、`c.DefaultQuery` |
| Path | `c.Param` |
| form-data / 文件上传 | `c.FormFile`、`c.PostForm`、`c.DefaultPostForm` |
| Header | `c.GetHeader` |
| Cookie | `c.Cookie` |

认证令牌和项目范围（`X-Project-Code`）由中间件从 Header 解析，API 从上下文取已校验结果，不重复解析，也不信任客户端传入的授权项目列表。

### Swagger 注释

对外 API 的 Swagger 注释至少写清：功能说明（`@Summary`）、请求参数（`@Param`，含来源）、响应结构（`@Success`/`@Failure`）、路由路径与方法（`@Router`）和鉴权要求（`Authorization` Header 或 `@Security`）。注释须与真实行为一致，现有写法见 `server/api/v1/system/audit.go`。新增接口必须带注释。照 gin-vue-admin 用 swaggo/swag 生成 `server/docs/` 并提供 `/swagger` 页面；接入与存量注释补齐见 [#73](https://github.com/Dcuhddrg/Optimus/issues/73)，接入前注释仍按上述要求维护。
