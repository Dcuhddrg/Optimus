# 模块说明索引

> 参照 gin-vue-admin `cee04b08` 的 `aiDoc/modules/module-index.md`，按 Optimus 现状改写。

## 模块文档要回答的问题

每个重要模块或插件用一份文档回答：

1. 这个模块做什么；
2. 后端入口在哪里；
3. 前端入口在哪里；
4. 依赖哪些数据或其他模块；
5. 对外暴露什么契约（接口、权限码、项目范围）；
6. 有哪些必须记住的限制。

聚焦职责和边界，不堆砌实现细节；代码事实按需通过图谱核对。

## 当前模块分组

| 分组 | 后端 | 前端 | 设计入口 |
| --- | --- | --- | --- |
| system | `server/{api/v1,service,model,router,source}/system/` | `web/src/view/{auth,system}/`、`web/src/api/system.ts` | [后端分层](backend-layer-rules.md)、[接口契约](../frontend-backend/boundary.md) |
| project | `server/{api/v1,service,model,router,source}/project/` | `web/src/view/project/`、`web/src/api/project.ts` | [多项目隔离](../superpowers/specs/2026-05-14-multi-project-cloud-isolation-design.md) |
| cloud | `server/{api/v1,service,model,router,source}/cloud/` | `web/src/view/cloud/`、`web/src/api/cloud.ts` | [云资源与项目](../memory/business/cloud-resources.md) |
| 运维中心 | 设计中 | `web/src/view/service/` | [运维中心](../memory/business/operations.md) |
| plugins | `server/plugin/<name>/`（尚无） | `web/src/plugin/<name>/`（尚无） | [插件开发约束](plugin-development.md) |
| deploy | `deploy/` | — | [项目画像](../relations/repo-profile.md) |

## 已有约束文档

- [backend-layer-rules.md](backend-layer-rules.md)：后端分层、装配入口、通用写法与迁移边界。
- [plugin-development.md](plugin-development.md)：前后端插件结构、入口职责与开发流程。

## 命名建议

新增模块文档时优先使用：

- `<module>.md`，如 `system.md`、`cloud.md`；
- `plugin-<name>.md`；
- `deploy-runtime.md`。

已有业务设计放在 `superpowers/specs/` 的，模块文档只链接，不复制正文。
