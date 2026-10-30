# 前端 API 示例

> 改写自 [gin-vue-admin / frontend/api-example.md](https://github.com/flipped-aurora/gin-vue-admin/blob/cee04b08d1262036ca33e3026e1e13d29471c7d5/aiDoc/examples/frontend/api-example.md)（`cee04b08`）。2026-10-30 适配 Optimus：保留上游讲解结构，代码、路径及现状说明已修改；上游许可见 [LICENSE.gin-vue-admin](../LICENSE.gin-vue-admin)。

## 这个文件负责什么

模块 API 负责路径、方法、参数编码和返回类型，公共请求层负责认证与 HTTP 错误；页面只调用模块函数。

## 什么时候应该这样写

- 新增模块请求。
- 分页、筛选或 CSV 下载。
- 修改字段或失败分类时核对全部调用方。

## 推荐写法示例

`web/src/api/system.ts` 的真实方法摘录，`auditQuery` 是同文件的参数编码助手：

```ts
export function listAuditLogs(filters: AuditLogFilter = {}, signal?: AbortSignal) {
  return apiRequest<AuditLogPage>(`/system/audit-logs${auditQuery(filters, true)}`, { signal });
}

export function exportAuditLogs(filters: AuditLogFilter = {}, signal?: AbortSignal) {
  return apiDownloadCSV(`/system/audit-logs/export${auditQuery(filters, false)}`, { signal });
}
```

`apiRequest` 固定前缀 `/api/v1`，返回原始 DTO，不自动解包 envelope。AuditLogPage 的真实字段是 `items/total/page/pageSize`；`actorUsername` 是已有快照。查询参数由 URLSearchParams 编码，导出复用已应用筛选、排除分页，并遵守现有最多 10000 条限制。

## 为什么这样写

取消信号贯穿到共享请求层；CSV 校验状态和 Content-Type，不把 JSON 错误保存为文件。登录使用公开请求配置，私有 401 沿用集中会话失效。

## 常见错误

- 从上游复制 `res.data.list` 到当前原始 DTO 接口。
- 页面手写 fetch 或拼认证头。
- 403 清理登录凭据、忽略 CSV 内容类型，或根据 data 字段猜协议。

## 真实参考文件

- [模块 API](../../../web/src/api/system.ts)、[请求层](../../../web/src/api/client.ts)、[真实类型](../../../web/src/types/system.ts)。
- [请求与导出回归](../../../web/src/api/system.test.ts)、[认证回归](../../../web/src/api/client.test.ts)。
