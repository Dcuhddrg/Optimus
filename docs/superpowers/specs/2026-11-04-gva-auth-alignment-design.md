# #56 认证与授权照 GVA 对齐设计

- 基线：gin-vue-admin `cee04b08d1262036ca33e3026e1e13d29471c7d5`；关联 [#56](https://github.com/Dcuhddrg/Optimus/issues/56)、[#55](https://github.com/Dcuhddrg/Optimus/issues/55)。
- 前置已合入：JWT 工具已在 `utils/jwt.go`（#60），响应协议已照 GVA（#61）。

## 1. 现状对照

| 环节 | GVA | Optimus 当前 |
| --- | --- | --- |
| 令牌传输 | 请求头 `x-token`，缺失时读 cookie `x-token`；登录与刷新时服务端写 cookie | 请求头 `Authorization: Bearer`；前端存 localStorage |
| 令牌校验 | `JWTAuth`：黑名单 → 解析 → 过期提示 → `c.Set("claims")` | 解析 → `c.Set("claims")`，无黑名单 |
| 临期刷新 | 剩余时间 < BufferTime 时签新令牌，响应头 `new-token`、`new-expires-at` 并写 cookie | 无；过期即需重新登录 |
| 登出/作废 | `jwt_blacklists` 表 + 内存缓存 `BlackCache`；启动时 `LoadAll` 载入缓存 | 登出只写审计，令牌在过期前仍可用 |
| 多点登录 | `UseMultipoint` 时 Redis 记录 用户→令牌，新登录作废旧令牌 | 无；Redis 连接无消费者（#55） |
| API 授权 | `utils.GetCasbin()` 按需创建 SyncedCachedEnforcer（gorm-adapter，表 `casbin_rule`）；`sub`=角色 ID，`obj`=去掉 RouterPrefix 的 URL.Path，`keyMatch2` 匹配 | `RBACService.AuthorizeAPI` 查 `sys_permissions`+`sys_role_permissions`；`CasbinHandlerWithPrefix` 把运行前缀映射到 `/api/v1` 规范路径 |
| 权限维护 | 菜单走 `sys_authority_menus`；API 策略由 `CasbinService.UpdateCasbin` 写 `casbin_rule` 后 `FreshCasbin` | 角色权限码（菜单+API 混合）由 `RoleService.UpdateRolePermissions` 写 `sys_role_permissions` |
| 遗留 | — | `initialize.Casbin()`、`global.OPTIMUS_ENFORCER`、`config.Casbin`、`config/casbin_model.conf`（Dockerfile 与 compose 均引用）无生产消费者 |

## 2. 目标方案

### 2.1 令牌（照 GVA）
- `utils/claims.go` 增加 `GetToken`/`SetToken`/`ClearToken`/`GetClaims`/`GetUserID`，语义同 GVA。
- `CustomClaims` 增加 `BufferTime`；`utils.JWT` 增加 `CreateTokenByOldToken`（GVA 用 singleflight 合并并发刷新）。
- `middleware.JWTAuth` 照 GVA：无令牌、在黑名单、过期、解析失败均 `NoAuth`+`ClearToken`；临期时写 `new-token`/`new-expires-at` 响应头与 cookie。
- 登录成功：照 GVA `TokenNext`，响应体返回 token 并写 cookie。
- 前端：请求头改为 `x-token`；收到 `new-token` 时更新本地令牌；401 处理不变。
- **影响**：发布后现有 Bearer 令牌全部失效，所有用户需重新登录一次。

### 2.2 黑名单（照 GVA）
- 新表 `jwt_blacklists`（`OPTIMUS_MODEL` + `Jwt text`），`global.BlackCache` 内存缓存，过期时间=令牌有效期。
- `initialize.OtherInit` 创建缓存；数据库可用后 `LoadAll` 载入。
- 现有 `POST /system/auth/logout` 改为：写审计 + `JsonInBlacklist` + `ClearToken`（保留 Optimus 路径，不新增 GVA 的 `/jwt/jsonInBlacklist`）。

### 2.3 授权（照 GVA Casbin）
- `utils/casbin_util.go`：`GetCasbin()` 同 GVA（SyncedCachedEnforcer + gorm-adapter，内置模型字符串）。
- `middleware.CasbinHandler`：`obj = TrimPrefix(URL.Path, RouterPrefix)`，`act = Method`；`sub` 依次取令牌中的每个角色 ID，任一允许即放行（**必要差异**：Optimus 用户可有多个角色，GVA 令牌只带当前 AuthorityId）。拒绝时 `FailWithDetailed(gin.H{}, "permission denied")`。删除 `CasbinHandlerWithPrefix` 前缀映射。
- 权限数据：`sys_permissions` 继续作为菜单+API 目录（对应 GVA 的菜单表与 `sys_apis`），API 路径改为不含前缀的规范形式（如 `/system/users/:id`）；
  - 角色的**菜单**权限仍存 `sys_role_permissions`（对应 `sys_authority_menus`），登录返回的 `menuPermissions` 不变；
  - 角色的**API** 权限存 `casbin_rule`（`p, <roleID>, <path>, <method>`）。`RoleService.UpdateRolePermissions` 按类型拆分写入，API 部分照 GVA `UpdateCasbin`（清除该角色策略 → 批量写入 → `FreshCasbin`），审计内容不变。
  - 种子：启动时把现有角色的 API 权限幂等迁移为 `casbin_rule`，之后删除 `sys_role_permissions` 中 API 类型的行；`admin` 核心权限保护规则保留。
- 删除 `RBACService.AuthorizeAPI` 及其表查询、`initialize.Casbin()`、`OPTIMUS_ENFORCER`、`config.Casbin`、`casbin_model.conf` 及 Dockerfile/compose 中的引用。
- 依赖：`casbin/casbin/v2` → `casbin/casbin/v3`，新增 `casbin/gorm-adapter/v3`（与 GVA 相同）。

### 2.4 不变
- 项目隔离 `ProjectScope`、LDAP 登录、审计事件、`X-Project-Code` 均不变。

## 3. 已确认决定（2026-11-04）

1. **cookie**：完全照 GVA，请求头 `x-token`，缺失时读 `x-token` cookie，登录与刷新时服务端写 cookie。
2. **多点登录与 Redis（#55）**：不实现 `UseMultipoint`；移除 Redis 初始化、`global.OPTIMUS_REDIS`、`system.use-redis` 与 `redis` 配置及依赖，#55 随之关闭。
3. **黑名单缓存**：照 GVA 引入 `github.com/songzhibin97/gkit` 的 `local_cache`。
4. **JWT 配置**：照 GVA 改为时长字符串（`expires-time: 7d`、`buffer-time: 1d`），用 `utils.ParseDuration` 解析；`config.yaml` 与部署配置同步修改。
5. **Go 版本**（2026-11-05）：gorm-adapter 自 v3.40 起才使用 casbin v3 且要求 Go 1.24，因此与 GVA 一致采用 casbin v3.10.0 + gorm-adapter v3.41.0，`go.mod` 与 `Dockerfile.backend` 升到 Go 1.24，gorm 随之升到 1.31。

## 4. 分批与验证

1. 令牌：传输、刷新、黑名单、登出、前端适配。
2. 授权：Casbin、数据迁移、角色权限写入、删除旧授权链和遗留配置。
3. 移除 Redis（决定 2），关闭 #55。

每批：后端 `go build/vet/test`，临时 MySQL 全量（`-p 1`）；前端 `tsc`、Vitest、`vite build`。重点用例：登录、令牌过期与刷新、登出后令牌失效、黑名单重启后仍生效、角色 API 权限增删即时生效、多角色、自定义 RouterPrefix、项目隔离与 401/code 7 语义。未覆盖：真实部署与浏览器端到端。
