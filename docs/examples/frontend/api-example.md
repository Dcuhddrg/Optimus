# 前端 API 示例

## 这个文件负责什么

模块 API 负责路径、方法、参数编码和返回类型，公共请求层负责认证与 HTTP 错误；页面只调用模块函数。

## 什么时候应该这样写

- 新增模块请求。
- 分页、筛选或 CSV 下载。
- 修改字段或失败分类时核对全部调用方。

## 推荐写法示例

`web/src/api/system.ts` 的真实方法摘录，`auditQuery`、`validAuditPage` 是同文件的参数编码和分页校验助手，`ApiError` 与显式适配器来自公共请求层：

```ts
export async function listAuditLogs(filters: AuditLogFilter = {}, signal?: AbortSignal) {
  const data = await apiRequestEnvelope<unknown>(`/system/audit-logs${auditQuery(filters, true)}`, { signal });
  if (!validAuditPage(data)) throw new ApiError('Invalid audit page data', 200, 'protocol');
  return data;
}

export function exportAuditLogs(filters: AuditLogFilter = {}, signal?: AbortSignal) {
  return apiDownloadCSV(`/system/audit-logs/export${auditQuery(filters, false)}`, { signal });
}
```

公共请求固定前缀 `/api/v1`。旧 `apiRequest` 返回原始 DTO，新 `apiRequestEnvelope` 只供明确迁移的模块选用，严格校验并返回 data，不猜测对象形状。审计 AuditLogPage 的真实字段是 `list/total/page/pageSize`；`actorUsername` 是已有快照。查询参数由 URLSearchParams 编码，导出复用已应用筛选、排除分页，并遵守现有最多 10000 条限制。

## 为什么这样写

取消信号贯穿到共享请求层；CSV 校验状态和 Content-Type，不把 JSON 错误保存为文件。登录使用公开请求配置，私有 401 沿用集中会话失效。

## 常见错误

- 把其他未迁移接口切到 envelope，或对已解包的审计结果再次读取 data。
- 页面手写 fetch 或拼认证头。
- 403 清理登录凭据、忽略 CSV 内容类型，或根据 data 字段猜协议。

## 真实参考文件

- [模块 API](../../../web/src/api/system.ts)、[请求层](../../../web/src/api/client.ts)、[真实类型](../../../web/src/types/system.ts)。
- [请求与导出回归](../../../web/src/api/system.test.ts)、[认证回归](../../../web/src/api/client.test.ts)。
