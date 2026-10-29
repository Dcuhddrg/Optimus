# 会话恢复修复实施计划

**Goal:** 修复 #27，保留非认证故障时的凭据并阻止未验证会话进入页面。

**Architecture:** 沿用请求封装、AuthContext 和 RequireAuth；显式区分公开登录请求，集中处理有效私有请求的 401，以可重试错误状态隔离恢复失败。

**Tech Stack:** React、TypeScript、React Router、Vitest、jsdom；不新增依赖。

## 步骤

- [x] 在 `web/src/api/client.test.ts` 与 `web/src/auth/AuthContext.test.tsx` 添加请求和真实 Provider/路由回归，先运行确认预期失败。
- [x] 修改 `web/src/api/client.ts`、`web/src/api/system.ts`：错误保留状态；公开登录不清理会话；旧 token/取消请求的 401 不注销当前会话。
- [x] 修改 `web/src/auth/AuthContext.tsx`、`RequireAuth.tsx`、`web/src/locale/messages.ts`：恢复错误、显式重试、守卫与中英文提示，清理失效恢复请求。
- [x] 运行定向回归并修复；审查 401 集中处理、非认证失败保留凭据、重试不放行及旧恢复响应隔离。
- [x] 执行前端测试、构建及 `git diff --check`，记录证据与未覆盖项，见下方结果。
- [x] 更新接口契约当前行为，并在 #27 记录实现、验证和交付进度；用户验收前保持 Issue 打开，提交/推送/合并另按授权执行。

## 本地验证（2026-10-29）

- 基线 `dd709e6`；最初 14 项回归均在旧实现上失败，同 token 登录与未决恢复竞争的补充回归也先失败再修复；最终认证定向 21 项通过。
- `pnpm` 命令不存在，`corepack pnpm` 启动报 `ERR_VM_DYNAMIC_IMPORT_CALLBACK_MISSING`。不修改依赖，直接使用已安装工具：`cd web && node node_modules/vitest/vitest.mjs run src/api/client.test.ts src/auth/AuthContext.test.tsx`。
- 构建：`cd web && node node_modules/typescript/bin/tsc -b && node node_modules/vite/bin/vite.js build`，通过。
- 全量：`cd web && node node_modules/vitest/vitest.mjs run`，35 项通过，1 个测试套件因 `approvalChains.test.ts` 引用不存在的 `approvalChains` 无法加载。将修改前 HEAD 的 web 归档到临时目录并复用依赖复验，同一套件失败、其余 14 项通过；未修改 #4 的审批链代码或测试。
- `git diff --check` 通过；文档检查 78 份 Markdown、211 个本地文件/章节链接通过。独立代码审阅未发现本批范围内阻塞问题。
- 未做真实认证联调、部署或浏览器端到端验收；不宣称登录与退出同时发起时的全局竞态已解决。
