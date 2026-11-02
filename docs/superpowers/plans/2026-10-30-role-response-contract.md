# 角色与权限响应迁移实施计划

> 按 superpowers:executing-plans 在当前会话执行；本批延续 #18 已确认的契约和批次方案。

**目标：** 六个角色/权限端点与角色管理调用同批迁移。

**架构：** 保留 Router → API → Service → Model，复用 response 和 apiRequestEnvelope；普通集合不伪造分页。

**技术栈：** Go/Gin、React/TypeScript、Vitest，不新增依赖。

## 1. 契约回归先行

- [x] 新增 `server/initialize/role_protocol_test.go`：真实 Router + RoleService + 可控 RoleStore，覆盖成功和业务/中间件失败，默认/自定义/根前缀，旧 LDAP 和探活。
- [x] 新增 `web/src/api/roleProtocol.test.ts` 和 `web/src/view/system/SystemRolesPage.test.tsx`，覆盖真实模块 API 与页面加载/保存。
- [x] 运行 `cd server && go test ./initialize -run TestRoleProtocol -count=1` 和前端定向 Vitest，确认因旧协议失败。

## 2. 最小迁移

- [x] 修改 `server/api/v1/system/role.go`：成功/错误使用 response，角色 list/total，权限 list，空集合不为 null；503 直接使用 response，不改变共享旧 helper。
- [x] 修改 `server/initialize/router.go`：登记六个方法/路径，复用共享参数错误的协议标记分支。
- [x] 修改 `web/src/api/system.ts`：三个函数显式使用 envelope 和准确 data 类型；修改 `web/src/view/system/SystemRolesPage.tsx` 读取 list。
- [x] 重跑定向测试，确认通过。

## 3. 验证与交付

- [x] 核对 Swagger 声明与六路实际响应；现有角色 handler 没有 Swagger 注释，本批补充集合/DTO envelope、失败状态和 401 null 的声明。
- [x] `cd server && go test ./...`、`go build ./...`、相关 race；数据库受阻时核实隔离数据库方案，不触碰应用数据。
- [x] `cd web && pnpm test`、`pnpm build`；pnpm 不可用时用现有 Node 直接运行已安装 Vitest/TypeScript/Vite，不安装新依赖。
- [x] 更新 `docs/frontend-backend/boundary.md` 的当前行为和本计划证据；`git diff --check`。
- [x] 依据 superpowers:requesting-code-review 独立复审，修复本批问题；交付保留 #18 后续 LDAP/project/cloud 与发布验收缺口。
- [x] 将本批交付范围、验证命令、结果和未覆盖项写入 Issue #18，保持 Issue 打开，不将本地验证当作发布或用户验收。

## 本批验证证据（2026-11-02）

- 新增前端 API 6 项和角色页面 1 项测试：先因旧协议失败，迁移后 7 项全部通过。真实 Router 测试已验证默认、自定义、根前缀的成功与失败分支。
- `GOCACHE=/tmp/optimus-go-build go test -race -count=1 ./initialize ./middleware ./utils/response`、`go build ./...` 通过。
- pnpm 不在 PATH，直接调用已有依赖：`node node_modules/vitest/vitest.mjs run` 有 87 项通过；既有 `approvalChains.test.ts` 因缺失 `./approvalChains` 无法加载。`git show HEAD:web/src/view/service/data/approvalChains.ts` 确认基线亦缺失，未扩范围修复。
- `node node_modules/typescript/bin/tsc -b && node node_modules/vite/bin/vite.js build` 通过；未新增或升级依赖。
- 独立代码复审通过；`git diff --check` 及 3 份 Markdown 的 15 个本地链接目标检查通过。
- 未部署；发布前仍须确认独立调用方、旧标签页及前后端发布顺序。#18 保持打开；LDAP、project、cloud 和真实环境验收仍待后续推进。

- 隔离临时 MySQL 全量 `OPTIMUS_TEST_MYSQL_DSN=<独立测试 DSN> GOCACHE=/tmp/optimus-go-build go test -count=1 -p 1 -json ./...`：262 项通过，0 失败、0 跳过（无测试文件的包单独列出，不计入测试跳过）。容器已停止并自动删除，未使用应用数据库。
- 本批候选范围与上述验证证据已写入 Issue #18；保持 Issue 打开，等待 PR 交付及用户验收。
