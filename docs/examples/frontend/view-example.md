# View 示例

## 这个文件负责什么

页面组织查询表单、表格/列表、详情与交互状态；不重写公共认证请求，也不把 UI 隐藏当成服务端授权。

## 什么时候应该这样写

- 新增后台页面。
- 接通搜索与分页。
- 补加载、空、错误与取消状态。

## 推荐写法示例

以下是按现有审计页面方式缩减的教学组件，不是新路由，也不包含完整筛选和分页 UI：

```tsx
import { useEffect, useState } from 'react';
import { listAuditLogs } from '../../api/system';
import { useI18n } from '../../locale/useI18n';
import type { AuditLog } from '../../types/system';

export function AuditListExample() {
  const { t } = useI18n();
  const [items, setItems] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    void listAuditLogs({ page: 1, pageSize: 20 }, controller.signal)
      .then(page => { if (!controller.signal.aborted) setItems(page.list); })
      .catch(() => { if (!controller.signal.aborted) setFailed(true); })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, []);

  if (loading) return <p role="status">{t('system.audit.loading')}</p>;
  if (failed) return <p role="alert">{t('system.audit.loadFailed')}</p>;
  if (items.length === 0) return <p>{t('system.audit.empty')}</p>;
  return <ul>{items.map(item => <li key={item.id}>{item.actorUsername || (item.actorUserId ? `#${item.actorUserId}` : t('system.audit.unknownActor'))}</li>)}</ul>;
}
```

完整的查询重试、草稿/已应用条件、分页重置、未知字段回退与详情见真实审计页。新页面应同步 `router/index.tsx`、导航配置、权限入口和中英文文案；不为文档注册此教学组件。

## 为什么这样写

查询区与展示区清楚，失败与空状态分开，取消旧请求防止覆盖新结果。沿用 CSS Modules 和主题变量，局部状态不强行放到 Context。

## 常见错误

- 一次请求加载首批记录却宣称完整列表。
- 将草稿条件用于导出，或筛选后不重置页码。
- 新页面只加路由，遗漏导航、权限和两种语言。

## 真实参考文件

- [完整审计页](../../../web/src/view/system/SystemAuditLogsPage.tsx)、[页面回归](../../../web/src/view/system/SystemAuditLogsPage.test.tsx)、[样式](../../../web/src/view/system/SystemAuditLogsPage.module.css)。
- [路由](../../../web/src/router/index.tsx)、[侧栏](../../../web/src/components/nav/SideNav.tsx)、[语言资源](../../../web/src/locale/messages.ts)。
