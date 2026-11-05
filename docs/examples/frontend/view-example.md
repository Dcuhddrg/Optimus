# View 示例

## 这个文件负责什么

`web/src/view/<模块>/<Name>Page.tsx` 负责查询、列表、表单和交互流程，样式放在同名 `*.module.css`。

## 什么时候应该这样写

- 新增管理页面
- 新增列表页 + 搜索 + 分页
- 新增创建或编辑表单

## 推荐写法示例

```tsx
import { FormEvent, useEffect, useState } from 'react';

import { listOrders } from '../../api/order';
import { PageHeader } from '../../components/page/PageHeader';
import { EmptyState } from '../../components/state/EmptyState';
import { useI18n } from '../../locale/useI18n';
import type { Order } from '../../types/order';
import styles from './OrdersPage.module.css';

export function OrdersPage() {
  const { t } = useI18n();
  const [keyword, setKeyword] = useState('');
  const [orders, setOrders] = useState<Order[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  async function load(nextPage = page) {
    setLoading(true);
    setError(null);
    try {
      const result = await listOrders({ page: nextPage, pageSize: 20, keyword });
      setOrders(result.list);
      setTotal(result.total);
      setPage(result.page);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('order.list.loadFailed'));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load(1);
  }, []);

  function handleSearch(event: FormEvent) {
    event.preventDefault();
    void load(1);
  }

  return (
    <section className={styles.page}>
      <PageHeader eyebrow={t('order.list.eyebrow')} title={t('order.list.title')} />

      <form className={styles.search} onSubmit={handleSearch}>
        <input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder={t('order.list.keyword')} />
        <button type="submit">{t('common.search')}</button>
      </form>

      {error ? <div className={styles.error}>{error}</div> : null}
      {!loading && orders.length === 0 ? (
        <EmptyState title={t('order.list.emptyTitle')} description={t('order.list.emptyDescription')} />
      ) : (
        <div className={styles.table}>
          {orders.map((order) => (
            <article key={order.id} className={styles.row}>
              <strong>{order.name}</strong>
            </article>
          ))}
        </div>
      )}
      <p className={styles.total}>
        {t('order.list.total')}: {total}
      </p>
    </section>
  );
}
```

## 为什么这样写

- 请求只通过模块 API 函数，页面处理加载、空数据、失败三种状态。
- 文案全部走 `useI18n()` 的 `t(key)`；key 受 `MessageKey` 类型约束，新增文案先在 [messages.ts](../../../web/src/locale/messages.ts) 补中英文，`t` 不做插值，数值单独渲染。
- 复用 `PageHeader`、`EmptyState`，样式用 CSS Module，不照搬上游 Element Plus / UnoCSS。

## 常见错误

- 页面里直接 `fetch` 或重复实现请求逻辑
- 硬编码中文或英文文案
- 组件过大，不拆查询区、列表区、表单区
- 忽略加载、空数据、失败或权限不足（`isPermissionDenied`）状态

## 真实参考文件

- [web/src/view/project/ProjectsPage.tsx](../../../web/src/view/project/ProjectsPage.tsx)
- [web/src/view/system/SystemAuditLogsPage.tsx](../../../web/src/view/system/SystemAuditLogsPage.tsx)
