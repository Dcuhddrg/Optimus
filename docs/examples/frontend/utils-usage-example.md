# 工具函数使用示例

> 改写自 [gin-vue-admin / frontend/utils-usage-example.md](https://github.com/flipped-aurora/gin-vue-admin/blob/cee04b08d1262036ca33e3026e1e13d29471c7d5/aiDoc/examples/frontend/utils-usage-example.md)（`cee04b08`）。2026-10-30 适配 Optimus：保留上游讲解结构，代码、路径及现状说明已修改；上游许可见 [LICENSE.gin-vue-admin](../LICENSE.gin-vue-admin)。

## 这个文件负责什么

优先复用已有公共能力，但不能把上游 utils 清单当成 Optimus 已安装的工具库。通用能力确有重复使用再抽取。

## 什么时候应该这样写

- 请求认证、错误或 CSV 下载。
- 操作名称、结果及资源类型本地化。
- 共享身份/语言上下文。

## 推荐写法示例

实际存在的函数调用形态（放在已有页面或组件上下文）：

```tsx
import { getAuditActionLabel, getAuditResultLabel } from '../../locale/systemLabels';
import { useI18n } from '../../locale/useI18n';

export function AuditLabelExample() {
  const { locale } = useI18n();
  const actionText = getAuditActionLabel('role.permissions.update', locale);
  const resultText = getAuditResultLabel('success', locale);
  return <span>{actionText} · {resultText}</span>;
}
```

HTTP 统一使用模块 API 和 `api/client.ts`；语言使用 useI18n，状态使用已有 Context。上游 `formatDate/CreateUUID/getDict/useBtnAuth` 不在这里假定可用。日期当前审计页使用浏览器 locale 格式化，不为单次使用新建工具包。

## 为什么这样写

复用实际存在的能力并链接实现，比编造 imports 更可靠。未知动作标签回退原始代码；菜单可见性不能代替 API 权限。

## 常见错误

- 引入不存在的上游别名路径或工具名。
- 为获取字典、按钮权限新增不必要依赖。
- 将未知动作强行翻译成已有业务含义。

## 真实参考文件

- [本地化标签](../../../web/src/locale/systemLabels.ts)、[语言 hook](../../../web/src/locale/useI18n.ts)。
- [公共请求层](../../../web/src/api/client.ts)、[会话上下文](../../../web/src/auth/AuthContext.tsx)、[完整使用示例](../../../web/src/view/system/SystemAuditLogsPage.tsx)。
