# ADR-003：前后端响应协议采用 gin-vue-admin

- 日期：2026-11-04。
- 状态：已由用户确认（“完全照 GVA”）；实施与验证见 [#18](https://github.com/Dcuhddrg/Optimus/issues/18)、[#27](https://github.com/Dcuhddrg/Optimus/issues/27)、[#28](https://github.com/Dcuhddrg/Optimus/issues/28)、#51、#56。
- 取代：#18 原定的“保留 HTTP 4xx/5xx、错误不改为 HTTP 200”，以及 [ADR-002](ADR-002-backend-structure-migration.md) 中“已确认的 HTTP 4xx/5xx 语义保留”一句。
- 现行规则：[接口契约](../../../frontend-backend/boundary.md)。

## 背景

#7 起草的接口契约（2026-10-27）在统一 `{code,data,msg}` 的同时保留 HTTP 4xx/5xx 表示错误，并计划分模块切换。后端结构按 gin-vue-admin（`cee04b08`）迁移后，响应协议与上游不一致，前端需要同时处理 HTTP 状态码和业务码两套失败语义。

## 决定

全部 JSON 接口采用 gin-vue-admin 的 `model/common/response` 协议：

- 业务结果一律 HTTP 200，`code` 0 为成功、7 为失败；参数错误、权限不足、资源不存在、冲突等都是 HTTP 200 + code 7。
- HTTP 401 只表示未认证（`NoAuth`）。
- 分页数据放在 `data` 内。
- 认证照 GVA：`x-token` 请求头与 cookie、`new-token` 刷新、登出黑名单、Casbin 策略存于 `casbin_rule`（#56）。
- 不保留按路由切换协议的开关，不再分模块切换。

## 例外

- CSV 成功响应保留 `text/csv` 附件，不包信封；失败仍返回 JSON。
- `/ping`、`/base/ping` 始终 200；`/health` 在数据库或启动数据初始化失败时返回 503 `{"status":"unhealthy"}`，与 GVA 不同（#51）。
- Recovery 中间件处理未捕获的 panic 时返回 HTTP 500，不带正文。

## 代价

- 前端不能再用 HTTP 状态码区分权限不足，改用 `isPermissionDenied()` 识别 `msg: "permission denied"`。
- 网关和监控按 HTTP 状态统计错误率时，看不到业务失败。
