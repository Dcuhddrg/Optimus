# 前后端协作与接口契约

全部 JSON 接口采用 gin-vue-admin（`cee04b08`）响应协议；决定、例外与取代关系见 [ADR-003](../adr/0003-gva-response-protocol.md)。

## 1. 适用范围与代码依据

以 [AGENT.MD](../../AGENT.MD) 为项目规则入口。本契约描述前后端可观察行为，不依赖 Service 内部目录；与代码不符时按[文档维护约定](../relations/documentation-workflow.md#状态与事实)核查登记。

| 内容 | 可定位依据 |
| --- | --- |
| 请求、认证 header、错误解析 | [utils/request.ts](../../web/src/utils/request.ts) |
| 集合类型与项目调用 | [system.ts 类型](../../web/src/types/system.ts)、[project.ts](../../web/src/api/project.ts)、[项目处理](../../server/api/v1/project/project.go)、[成员处理](../../server/api/v1/project/member.go) |
| 系统接口与审计导出 | [system.ts API](../../web/src/api/system.ts)、[audit.go](../../server/api/v1/system/audit.go) |
| 云资源分页 | [cloud.ts](../../web/src/api/cloud.ts)、[asset.go](../../server/api/v1/cloud/asset.go) |

## 2. 前后端职责与协作

参考 gin-vue-admin 的 [边界说明](https://github.com/flipped-aurora/gin-vue-admin/blob/main/aiDoc/frontend-backend/boundary.md)、[前端约束](https://github.com/flipped-aurora/gin-vue-admin/blob/main/aiDoc/frontend-backend/frontend-rules.md)和[工具函数约束](https://github.com/flipped-aurora/gin-vue-admin/blob/main/aiDoc/frontend-backend/frontend-utils.md)，结合 Optimus 的 React + TypeScript 技术栈执行；不引入上游 Vue、Pinia 或其专用组件库。

| 责任方 | 职责 |
| --- | --- |
| 后端 | Router 注册与中间件，API 绑定和校验请求、转换响应，Service 实现业务规则与授权相关逻辑，Model 定义数据结构；服务端是权限及项目隔离的最终执行方 |
| 前端 | 路由和页面流程、表单交互、展示与本地状态；通过模块 API 和明确类型对接服务端，权限控制用于改善体验，不代替服务端校验 |
| 共同 | 以请求、响应、字段、错误及权限约束为接口契约；变更时核对实现、调用方和 Swagger，分清既有行为与目标 |

需求涉及接口时先确定路径、方法、请求参数、响应 DTO、HTTP/业务错误、认证与项目范围；破坏性变更还要列出受影响调用方、发布顺序与回退方式。Swagger 与实际行为冲突时以实测和代码为当前事实，修正不一致后再作为接口依据。

## 3. 当前接口形状

所有 JSON 接口返回 `{code,data,msg}`，下表只列 data 的形状。

| 场景 | data |
| --- | --- |
| 登录、当前用户 | 原有登录/用户 DTO |
| 用户列表 | `{list,total}`，完整集合，不分页 |
| 角色列表 | `{list,total}`，完整集合，不分页 |
| 权限树 | `{list}` |
| 项目、成员列表 | `{list}` |
| 云账号列表 | `{list}`，支持标签筛选，不分页 |
| 云资产、审计查询 | `{list,total,page,pageSize}` |
| 登出、删除等无业务数据操作 | `{}` |
| CSV 导出、health/ping | 例外，见[第 7 节](#7-csv探活与无正文例外) |

前端只有一个 JSON 请求函数 `apiRequest`（[utils/request.ts](../../web/src/utils/request.ts)）：校验 code/data/msg，code 为 0 时返回 data，否则抛出 `kind: 'business'` 的 `ApiError`；HTTP 非 2xx 抛出 `kind: 'http'`。不通过 `body.data ?? body` 猜测协议，data 可能为 null、0 或 false。

## 4. 响应协议与 HTTP 语义

与 gin-vue-admin 的 `model/common/response` 一致：

| 字段 | 约定 |
| --- | --- |
| code | 整数；成功为 0（`SUCCESS`），失败为 7（`ERROR`） |
| data | 必须存在；成功返回已声明的 DTO/集合/分页数据；无业务数据时为 `{}`；`NoAuth` 为 `null`；`FailWithDetailed` 可携带错误详情 |
| msg | 字符串，供人阅读；不暴露凭据及内部 SQL 等细节 |

成功示例：

```json
{"code":0,"data":{"id":1,"name":"示例项目"},"msg":"成功"}
```

失败示例（HTTP 200）：

```json
{"code":7,"data":{},"msg":"no access to project"}
```

未认证示例（HTTP 401）：

```json
{"code":7,"data":null,"msg":"invalid or expired token"}
```

- **业务结果一律 HTTP 200**：参数错误、权限不足、资源不存在、冲突、服务不可用等都是 HTTP 200 + code 7。后端用 `Ok`/`OkWithMessage`/`OkWithData`/`OkWithDetailed`/`Fail`/`FailWithMessage`/`FailWithDetailed`，不直接 `c.JSON` 写业务结果。
- **HTTP 401 只表示未认证**：JWT 缺失、格式错误或过期时由 `NoAuth` 返回。JWT 中间件每个请求都会读取用户的当前状态和角色（[用户管理设计 3.1](../design/2026-11-06-user-management-design.md#31-权限即时生效d1)），账号已停用返回 `msg: "account is disabled"`、账号不存在返回 `msg: "account does not exist"`，同样是 401；前端据 `account is disabled` 在登录页提示账号已停用。
- **HTTP 500 只表示未处理的 panic**：Recovery 中间件中止请求，不返回正文。`/health` 在启动降级时返回 503，见第 7 节。
- 前端：HTTP 200 且 code 0 才算成功；HTTP 200 + code 7 是业务失败；JSON 解析失败或字段缺失是协议错误，不冒充正常数据。
- 权限不足由 Casbin 中间件返回 `msg: "permission denied"`；Casbin 策略重载失败、尚未恢复期间返回 `msg: "permission system is syncing, please retry later"`（HTTP 200 + code 7，不注销）；前端用 `isPermissionDenied()` 识别，不再依赖 HTTP 403。其他分支不依赖 msg 文案。

## 5. 集合、分页与字段

分页数据放在响应的 **data 内**，不是与 code/msg 平级：

```json
{"code":0,"data":{"page":1,"pageSize":20,"total":0,"list":[]},"msg":"成功"}
```

- 真正分页的接口返回 page、pageSize、total、list；total 为完整筛选结果数，不是当前页长度，空页返回 list=[]，保留真实 total。
- page 从 1 开始，默认 pageSize=20，范围 1～200，非法值回退默认（`model/common/request.PageInfo.Normalize`）。
- 普通集合建议返回 `data: { list: [] }`；业务需要计数时可声明真实 total，但不伪造 page/pageSize。权限树和非分页项目集合不能只为结构统一强行分页化。
- 新增 JSON 字段使用 camelCase；现有字段名和类型保持调用方一致，重命名须单独设计并同步所有调用方，不全局批量替换。
- 集合为空返回数组，不用 null 代替；可选与 nullable 分开声明，不把字段缺失、null、空字符串混为一种状态。普通无业务数据与未认证响应按第 4 节区分；不可将 null 泛化为所有失败响应。
- 时间使用带时区的 RFC3339/ISO 字符串，前端负责展示本地化；现有 ID 类型先保持不变，超出 JavaScript 安全整数范围等问题需单独明确，不能默默截断或改类型。

## 6. 请求、认证与权限协作

- view 通过 `web/src/api/` 的模块函数发请求；目前公共封装固定 `/api/v1`，后端 RouterPrefix 的自定义能力不等于前端 base URL 已自动适配。
- JSON 请求使用 Content-Type: application/json，GET 参数使用明确的查询类型；区分未定义字段、空字符串和 undefined。通用请求层不自动重试非幂等写操作。
- 认证照 GVA：令牌放在 `x-token` 请求头，缺失时读 `x-token` cookie（登录与刷新时服务端写入）；剩余有效期小于 buffer-time 时响应头 `new-token` 返回新令牌，前端替换本地令牌；登出把令牌加入黑名单。X-Project-Code 只是资源项目选择，不是授权凭证。后端执行 JWT、Casbin（`casbin_rule`，按角色 ID、去掉路由前缀的路径和方法）、成员资格和项目隔离检查。云资源使用项目范围，系统和项目恢复入口不受无关的过期项目选择拦截；项目管理按路径项目独立鉴权。
- 携带当前凭据且未取消的私有请求明确 401 才触发凭据失效；登录表单 401 显示认证失败，不无条件注销另一会话；权限不足和失效的项目选择（HTTP 200 + code 7）不注销用户，失效项目选择提示重新选择。网络错误、5xx 或解析异常不直接证明 token 失效：保留凭据并提供重试，但不将未确认状态显示为认证成功。会话恢复设计见 [session recovery](../design/2026-10-29-session-recovery-design.md)。

## 7. CSV、探活与无正文例外

- CSV 成功响应保留 text/csv、Content-Disposition 和文件名，不包 `{code,data,msg}`。下载适配器检查 Content-Type：JSON 表示失败（HTTP 200 + code 7 或 401），不能当文件保存。审计导出复用已应用筛选并排除分页参数，上限 10000 条，范围见[审计查询设计](../design/2026-10-29-audit-query-experience-design.md)。
- health/ping 保留小响应：`/ping`、`/base/ping` 始终 200；`/health` 正常时 200 `{"status":"ok"}`，数据库或启动数据初始化失败时 503 `{"status":"unhealthy"}`（与 GVA 不同的明确例外，见 ADR-003）。
- 其他上传、流式或非 JSON 接口存在时，必须独立声明内容类型及错误方式。
- 已声明的 JSON 操作不通过省略整个响应正文代替 data；若保留 204 等无正文响应，应明确列为例外并使用相应适配器。

## 8. 前端开发规则

页面、组件、状态、样式和工具复用见 [前端开发与复用规范](frontend-rules.md)。

## 9. 新增或修改接口

1. 后端 API 只用 `response` 包的 GVA 函数输出 JSON；中间件失败同样使用这些函数并 `c.Abort()`。不新增按路由切换协议的开关。
2. 前端模块函数统一调用 `apiRequest` 并声明 data 的类型；页面通过 `ApiError.kind`/`code` 区分网络、协议、业务失败。
3. 修改 DTO 或 data 形状时，同批修改后端 API、前端 API/类型/页面、本文档第 3 节和测试；存在独立调用方时先设计兼容与回退。

## 10. 文档与代码验证职责

- 后端同步请求/响应 DTO、处理逻辑、HTTP/业务码及 Swagger；前端同步 API、类型、错误处理和调用方。核对真实路由、方法、参数、字段及页面状态，明确兼容和回退方案的负责人及退出条件。
- 使用真实路径和调用链验证登录/当前用户、权限与项目隔离、集合与分页、空数据、HTTP 非 2xx、业务和协议失败、CSV/探活及新旧调用兼容；页面变更检查主要操作、失败和空状态，不能只测试自造返回值。
- 接口变更时同步本文档；实施与验证证据记在对应 Issue/PR，不写入本文。文档检查不能代替接口验证；按[工程操作参考](../relations/engineering-workflow.md#验证步骤)运行受影响的测试和构建，记录未覆盖或受阻场景。
