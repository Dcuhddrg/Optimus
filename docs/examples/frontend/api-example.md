# 前端 API 示例

## 这个文件负责什么

`web/src/api/<模块>.ts` 把后端接口封装成带类型的函数，统一调用 [client.ts](../../../web/src/api/client.ts) 的 `apiRequest`。页面不直接 `fetch`。

## 什么时候应该这样写

- 新增模块接口
- 给页面提供列表、详情、创建、更新、删除方法
- 为插件页面补接口封装

## 推荐写法示例

```ts
import { apiRequest } from './client';
import type { CreateOrderPayload, Order } from '../types/order';

type OrderList = { list: Order[]; total: number; page: number; pageSize: number };

// GET /orders：分页查询
export function listOrders(params: { page: number; pageSize: number; keyword?: string }) {
  const query = new URLSearchParams({ page: String(params.page), pageSize: String(params.pageSize) });
  if (params.keyword) query.set('keyword', params.keyword);
  return apiRequest<OrderList>(`/orders?${query}`);
}

// POST /orders：创建
export function createOrder(payload: CreateOrderPayload) {
  return apiRequest<Order>('/orders', { method: 'POST', body: JSON.stringify(payload) });
}

// DELETE /orders/:id：无业务数据时 data 为 {}
export function deleteOrder(id: number) {
  return apiRequest<Record<string, never>>(`/orders/${id}`, { method: 'DELETE' });
}
```

## 为什么这样写

- `apiRequest` 统一处理 `/api/v1` 前缀、`x-token` 与 `X-Project-Code` 请求头、`new-token` 刷新和 401 会话失效。
- 它按 GVA 协议校验 `{code,data,msg}`：code 为 0 返回 `data`，否则抛出 `kind: 'business'` 的 `ApiError`；调用方拿到的就是 `data`，不再判断 `res.code`。
- 泛型写明 `data` 的形状，与[接口契约](../../frontend-backend/boundary.md#3-当前接口形状)第 3 节一致。

## 常见错误

- 在页面里直接 `fetch` 或自己拼认证头
- 在调用侧再判断 `res.code === 0`（`apiRequest` 已解包）
- 用类型断言掩盖与真实响应不符的类型
- URL、方法、参数位置与后端路由不一致

## 真实参考文件

- [web/src/api/project.ts](../../../web/src/api/project.ts)
- [web/src/api/cloud.ts](../../../web/src/api/cloud.ts)
- [web/src/api/client.ts](../../../web/src/api/client.ts)
