# 共享状态示例（Context）

## 这个文件负责什么

跨页面共享的状态和异步动作放在 React Context，并通过 `useXxx` hook 读取；对应上游的 Pinia store。只在一个页面用的状态留在页面组件里。

## 什么时候应该这样写

- 当前用户、当前项目等多页面共享的状态
- 多个页面都要触发的重新加载动作
- 需要统一持久化（如 localStorage）的选择

## 推荐写法示例

```tsx
import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react';

import { listOrders } from '../api/order';
import type { Order } from '../types/order';

interface OrderContextValue {
  orders: Order[];
  loading: boolean;
  error: string | null;
  hasData: boolean;
  reload: () => Promise<void>;
}

const OrderContext = createContext<OrderContextValue | null>(null);

export function OrderProvider({ children }: { children: ReactNode }) {
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const hasData = useMemo(() => orders.length > 0, [orders]);

  async function reload() {
    setLoading(true);
    setError(null);
    try {
      const result = await listOrders({ page: 1, pageSize: 20 });
      setOrders(result.list);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load orders');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void reload();
  }, []);

  return (
    <OrderContext.Provider value={{ orders, loading, error, hasData, reload }}>{children}</OrderContext.Provider>
  );
}

export function useOrders() {
  const value = useContext(OrderContext);
  if (!value) {
    throw new Error('useOrders must be used inside OrderProvider');
  }
  return value;
}
```

## 为什么这样写

- Provider 集中加载、loading 和 error，页面只调用 `useOrders()`。
- hook 在 Provider 外使用时直接报错，避免静默拿到空状态。
- 错误信息来自 `ApiError.message`（即后端 `msg`），网络、协议、业务失败由 `ApiError.kind` 区分。

## 常见错误

- 把只属于一个页面的表单、弹窗状态放进全局 Context
- 在 Provider 里操作 DOM
- 不处理 loading / error，页面状态混乱
- 为状态管理引入新依赖（Pinia、Redux 等）而不先确认

## 真实参考文件

- [web/src/auth/ProjectContext.tsx](../../../web/src/auth/ProjectContext.tsx)
- [web/src/auth/AuthContext.tsx](../../../web/src/auth/AuthContext.tsx)
