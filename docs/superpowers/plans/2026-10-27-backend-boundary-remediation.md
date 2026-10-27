# 后端路由与授权边界整改 Implementation Plan

**Goal:** 在保持现有调用契约和隔离语义下，修复项目作用域过宽及路由前缀与权限键不一致。

**Architecture:** 沿用 Router → API → Service → Model。仅收敛中间件装配及权限查询前缀，不迁移服务内部目录。设计见 [首批规格](../specs/2026-10-27-backend-boundary-remediation-design.md)。

**Tech Stack:** Go、Gin、现有 testify/httptest 与 MySQL 测试夹具，无新增依赖。

**状态：** 首批实施与技术验证记录。用户已确认 #12/#13 验收通过；验收和后续进度以 GitHub Issues 为准，#11 持续跟踪后续批次。

## 文件所有权

- 根代理：`server/initialize/router.go`、新增 `server/initialize/router_project_scope_test.go`。
- 前缀修复代理：`server/middleware/casbin_rbac.go`、`server/middleware/casbin_rbac_test.go`、新增 `server/initialize/router_prefix_test.go`。
- 文档：本规格/计划、相关能力证据；#6 的 4 份文档保留，并按用户后续授权与本批代码分开提交。

## Task 1：基线与登记

- [x] 图谱核对模块边界，登记 #12/#13 并在 #11 中排期。
- [x] `go test -json ./...` 记录未配置 MySQL 的跳过项；缓存写入受限后提权重跑。
- [x] 隔离 MySQL 基线：`go test -count=1 -p 1 -json ./...` 149 通过，无失败/跳过；`go build ./...` 通过。

## Task 2：自定义前缀授权（#13）

- [x] 写现有 Routers 集成回归并观察 RED：默认前缀通过，自定义/根/尾斜杠/参数路由错误拒绝。
- [x] 最小中间件路径映射，保留旧构造器；验证前缀边界和未知权限拒绝。
- [x] 根代理用 privateGroup.BasePath()、defaultRouterPrefix 接入装配，运行集成测试观察 GREEN。

## Task 3：项目范围收敛（#12）

- [x] 添加现有 Routers 的行为回归，观察失效项目选择错误阻断系统及项目列表的 RED。
- [x] ProjectScope 移至云资源组；核对项目管理按路径授权不依赖全局选中项目。
- [x] 断言云资源仍拒绝失效选择、系统不再耦合项目解析错误、JWT/RBAC 及路径项目越权拒绝保持生效。

## Task 4：整体验证与交付

- [x] `go test ./middleware ./initialize -count=1`、`go test -race ./middleware ./initialize` 通过。
- [x] 隔离 MySQL 全量测试及 `go build ./...` 通过，记录测试数量、跳过及残余风险。
- [x] 独立审阅、文档链接和 `git diff --check`；更新 #2/#11/#12/#13，子项保持待用户验收，不自动关闭父项。

## 验证记录（2026-10-27）

- #12 回归 RED：旧装配让失效选择下的系统及项目列表返回 403/404，让系统接口受项目解析故障影响返回 500；修复后预期行为通过。
- #13 集成 RED：非默认前缀下原权限路径匹配失败；接入规范路径映射后默认、自定义、根、尾斜杠、参数路由及拒绝场景通过。
- 隔离 MySQL 全量命令为 `go test -count=1 -p 1 -json ./...`：基线 149，通过整改后 190（含子测试），均无失败或用例跳过；15 个无测试文件包不计为跳过用例。
- `go test -race ./middleware ./initialize`、`go build ./...` 通过；独立审阅未发现重要问题，并独立重跑针对性路由/中间件测试通过。
- MySQL 测试使用独立 `optimus_issue11_*` 库并串行跑包，避免原有清表夹具互相影响；应用库、部署配置和依赖保持不变。
- 未运行前端构建、真实云联调和部署验收；接口形状迁移留待 #7。用户已确认 #12/#13 验收通过并授权提交、推送及创建 PR；两项在 GitHub 完成关闭，父项 #11 继续跟踪剩余范围，交付提交及 PR 在 Issue 中关联。
