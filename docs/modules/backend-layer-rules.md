# 后端分层与迁移规则

> 项目规则以 [AGENT.MD](../../AGENT.MD) 为准；三模块结构迁移及验收见已关闭的 [#11](https://github.com/Dcuhddrg/Optimus/issues/11)。

## 后端分层契约

新模块与存量模块统一遵循以下分层、目录和依赖方向：

```text
Router -> API -> Service -> Model
```

| 层 | 职责与位置 | 依赖边界 |
| --- | --- | --- |
| Router | `server/router/<module>/`，注册路径、中间件与 public/private 分组 | 连接 API，不承担业务处理 |
| API | `server/api/v1/<module>/`，绑定和校验 HTTP 输入、调用服务、组织响应 | 通过 Service 处理业务，不绕过服务直接访问持久化 |
| Service | `server/service/<module>/`，业务规则、流程编排和数据访问协作 | 使用 Model 及所需基础设施，不依赖 `gin.Context`，不反向调用 API/Router |
| Model | `server/model/<module>/`，实体、请求/响应数据结构和持久化映射 | 不反向依赖 Router/API/Service |

- 保留现有 `enter.go` 分组注册和组合入口，新增分组沿用对应模块的注册方式；不得绕过已有入口另建平行注册链。
- `main` 与 `initialize` 负责启动和模块装配；随模块迁移同步调整导入与注入，保留必要能力，不保留重复的平行装配入口。
- Service 不接收 `gin.Context`；HTTP 处理留在 API，所需业务参数显式传入。Swagger 注释与真实接口行为保持一致。
- 新旧模块均按上述四层职责组织；具体开发指引在 [#8](https://github.com/Dcuhddrg/Optimus/issues/8) 维护，样例应使用已迁移并验证的模块。

### 后端结构维护要求

- 不得重新引入已清理的旧 `usecase/infra` 分层；新发现的结构差异在 GitHub Issue 跟进。
- 业务流程与服务内部接口归入 `server/service/<module>/`；实体、请求/响应数据结构与持久化映射归入 `server/model/<module>/`。数据访问实现位于对应 Service，不能把业务流程搬进 Model。
- 外部 SDK、协议适配和公共能力按职责迁入明确的模块或插件位置，每项给出旧路径、新路径、调用者与测试映射；避免仅给旧层改名后保留平行体系。
- 同批调整 API 导入、服务构造、`enter.go`、初始化注入和测试。必要接口可保留用于解耦与测试，不能成为旧包长期存在的豁免理由。
- 结构调整保持已有接口返回、HTTP 状态、权限和项目隔离行为；响应协议改造由 #18 独立实施。
- 结构调整须清除临时兼容入口，并按影响范围验证构建和隔离 MySQL 回归；进度与剩余差异在 GitHub 跟进。

2026-10-27 的盘点曾发现 system/project/cloud 旧目录；截至 2026-10-28，三批结构迁移已合入并由用户验收，入口清理已合入，#11 已关闭。现行决定及旧决定的替代关系见 [ADR-002](../memory/decisions/ADR-002-backend-structure-migration.md)。

## 装配与注册入口

`initialize/router.go` 创建各模块并注入 API handlers；`router/enter.go` 汇总模块 RouterGroup，各模块 `router/*/enter.go` 负责注册路由。当前 Service 通过各模块构造函数注入，API 使用模块 handlers；未使用的全局 ServiceGroup 已移除。后续如需统一 API/Service 的标准注册形式，应先明确设计并验证装配链，避免重复注册。
