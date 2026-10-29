# 审计日志查询体验实施计划

**Goal:** 完成 #28 分页及已确认第一批审计阅读、查询和导出优化。

**Architecture:** 沿用系统模块 API、公共请求认证能力和页面 CSS；已应用筛选控制分页和导出，详情只展示现有数据。

**Tech Stack:** React、TypeScript、CSS Modules、Vitest、jsdom；不新增依赖。

- [x] 在 `web/src/api/system.test.ts` 和 `web/src/view/system/SystemAuditLogsPage.test.tsx` 编写真实请求/页面回归，先确认旧实现失败。
- [x] 在 `web/src/types/system.ts` 声明真实审计分页及筛选类型，补操作人快照；更新 `api/system.ts` 编码查询和导出。
- [x] 在 `api/client.ts` 复用认证、项目 header、HTTP/网络错误和 #27 失效规则，提供 CSV 下载适配；回归原认证测试。
- [x] 更新 `SystemAuditLogsPage.tsx`、对应 CSS、`locale/systemLabels.ts` 和 `locale/messages.ts`；完成分页、已应用筛选、详情、加载/失败/空、重试及 CSV 下载。
- [x] 定向测试与全量前端测试、TypeScript 和 Vite 构建；文档链接及 `git diff --check`，独立审阅。
- [x] 在 #28 记录实际结果、未覆盖和后续记录质量范围；用户验收前不关闭，提交发布按后续授权。

## 本地验证（2026-10-29）

- 基线 `536f8b3`；最初 16 项审计回归在旧实现上失败，新增 API 与界面功能后通过；补充页码越界、取消导出及旧 token/取消 CSV 401 用例后，本批新增 20 项通过。
- `cd web && node node_modules/vitest/vitest.mjs run src/api/system.test.ts src/view/system/SystemAuditLogsPage.test.tsx src/api/client.test.ts src/auth/AuthContext.test.tsx`：41 项通过，包含原 #27 的 21 项回归。
- `cd web && node node_modules/typescript/bin/tsc -b && node node_modules/vite/bin/vite.js build`：通过。pnpm 入口不可用，使用仓库已安装工具，不新增依赖。
- 全量 `cd web && node node_modules/vitest/vitest.mjs run`：55 项通过，1 个既有审批链套件无法加载（缺 `approvalChains`）。在干净的 HEAD 归档复验，同一套件失败、原有 35 项通过；不修复 #4 范围。
- 真实 Chrome 加载生产构建，以模拟 fetch 提供假账号和审计记录，中文页面加载、第二页及展开详情冒烟通过；截图 `/tmp/optimus-audit-desktop.png`。本轮浏览器只连接隔离的本机静态服务，未修改 Compose 或运行中的应用，不构成真实后端认证、审计持久化或部署验收。
- 设计/计划及独立代码审阅通过。审阅发现的 CSS 空白格式已清理，`git diff --check` 通过。
- 后端记录质量另见 #36；本批未修改数据库、后端、依赖或部署配置，尚未提交、推送或发布，用户验收后再交付结案。

## 用户验收

2026-10-29，用户反馈“嗯，部署了看，没问题”，确认第一批体验，并随后授权提交 PR、合入 main。此前“本地验证”保留当时记录；最终提交、PR 与交付状态以 #28 为准，不扩大为助手独立验证真实认证或持久化链路。
