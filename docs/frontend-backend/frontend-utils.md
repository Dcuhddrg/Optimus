# 前端工具函数约束

> 参照 gin-vue-admin `cee04b08` 的 `aiDoc/frontend-backend/frontend-utils.md`，按 Optimus 的 React + TypeScript 现状改写。

## 核心原则

开发前端功能前，先检查并复用 `web/src/utils/` 和下表中的已有能力，不重复造轮子。GVA 的工具清单不代表 Optimus 已有对应实现，以本页和源码为准。

## 现有能力

| 位置 | 导出 | 用途 |
| --- | --- | --- |
| [utils/request.ts](../../web/src/utils/request.ts) | `apiRequest` | JSON 请求统一入口：带令牌和项目头，解析 `{code,data,msg}`，刷新令牌 |
| | `apiDownloadCSV` | CSV 下载；失败时按 JSON 信封解析错误 |
| | `ApiError`、`isPermissionDenied` | 错误分类（`http`/`network`/`protocol`/`business`）与无权限判断 |
| | `get/set/clearStoredToken`、`get/setStoredProjectCode`、`setUnauthorizedHandler` | 令牌与当前项目存储、401 处理；只由 auth 层调用 |
| [utils/format.ts](../../web/src/utils/format.ts) | `formatDateTime` | 日期时间格式化，空值返回 fallback |
| [auth/](../../web/src/auth/AuthContext.tsx) | `useAuth`、`useProject`、`RequireAuth` | 当前用户、菜单权限 `hasMenu`、当前项目、路由守卫 |
| [locale/](../../web/src/locale/useI18n.ts) | `useI18n`、`systemLabels` 的 `getXxxLabel` | 文案翻译与状态、角色、权限、审计枚举的显示名 |
| [components/](../../web/src/components/page/PageHeader.tsx) | `PageHeader`、`EmptyState`、`TagChip`、`TagEditor`、`TagFilter`、`ProjectSwitcher` | 页面骨架、空状态、标签、项目切换 |

用法示例见[复用公共能力示例](../examples/frontend/reuse-example.md)。

## 强制使用场景

- 发起 HTTP 请求：在 `web/src/api/` 中调用 `apiRequest` 或 `apiDownloadCSV`，view 不直接 `fetch`，不自行拼认证头或项目头。
- 判断无权限：用 `isPermissionDenied`，不比对错误文案。
- 显示日期时间：用 `formatDateTime`，不在页面里各写一套 `toLocaleString`。
- 显示枚举（状态、角色、权限、审计动作）：用 `systemLabels` 的 `getXxxLabel`，不在页面里写映射表。
- 菜单或页面权限：用 `useAuth().hasMenu(code)`。Optimus 没有按钮级权限数据，不引入 GVA 的 `btnAuth`。
- 跨组件共享状态：用已有 Context（`AuthContext`、`ProjectContext`、`LanguageProvider`），不引入 Pinia、事件总线或新的全局状态库。

## 新增工具

- 只有两个以上调用方确实需要时才提取到 `web/src/utils/`，单次使用留在模块内。
- 新工具写 TypeScript 类型和单测（同目录 `*.test.ts`），并更新本页表格。
- GVA 有而 Optimus 没有的能力（字典、UUID、命名转换、图片压缩、动态路由等），按需求设计后再加，不按上游清单预先补齐。
