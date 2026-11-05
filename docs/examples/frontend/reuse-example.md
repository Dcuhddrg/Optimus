# 复用公共能力示例

## 这个文件负责什么

需要通用能力时先复用已有实现，不另造一套。Optimus 没有上游的 `web/src/utils/`；公共能力分布在下表位置。

| 能力 | 位置 |
| --- | --- |
| HTTP 请求、令牌、项目头、错误分类 | [web/src/api/client.ts](../../../web/src/api/client.ts)：`apiRequest`、`apiDownloadCSV`、`ApiError`、`isPermissionDenied` |
| 当前用户、当前项目 | [web/src/auth/](../../../web/src/auth/AuthContext.tsx)：`useAuth`、`useProject` |
| 多语言 | [web/src/locale/useI18n.ts](../../../web/src/locale/useI18n.ts)：`useI18n().t` |
| 页面骨架与状态 | [web/src/components/](../../../web/src/components/page/PageHeader.tsx)：`PageHeader`、`EmptyState`、`TagChip`、`TagFilter` 等 |

## 推荐写法示例

```tsx
import { apiDownloadCSV, isPermissionDenied } from '../../api/client';
import { useProject } from '../../auth/ProjectContext';
import { PageHeader } from '../../components/page/PageHeader';
import { useI18n } from '../../locale/useI18n';

export function OrderExportBar() {
  const { t } = useI18n();
  const { currentProject } = useProject();

  async function handleExport() {
    try {
      const blob = await apiDownloadCSV('/orders/export');
      // 保存 blob ……
    } catch (error) {
      alert(isPermissionDenied(error) ? t('order.export.forbidden') : t('order.export.failed'));
    }
  }

  return <PageHeader title={currentProject?.name ?? t('project.switcher.empty')} actions={<button onClick={() => void handleExport()}>{t('order.export.action')}</button>} />;
}
```

## 为什么这样写

- 权限不足按 GVA 协议是 HTTP 200 + code 7，统一用 `isPermissionDenied` 判断，不检查 HTTP 403。
- CSV 下载用 `apiDownloadCSV`，失败时的 JSON 响应已由它识别。
- 当前项目从 `useProject` 读取，`X-Project-Code` 由 `apiRequest` 自动带上。

## 常见错误

- 自己写 `fetch` 或下载逻辑，绕开 `client.ts`
- 用 `error.status === 403` 判断权限
- 手写多语言或在组件里存一份当前项目
- 只用一次的逻辑急于抽成新工具或新依赖

## 真实参考文件

- [web/src/api/client.ts](../../../web/src/api/client.ts)
- [web/src/view/system/SystemAuditLogsPage.tsx](../../../web/src/view/system/SystemAuditLogsPage.tsx)
- [web/src/locale/useI18n.ts](../../../web/src/locale/useI18n.ts)
