# 后端按 gin-vue-admin 组织方式迁移计划

> 执行时使用 superpowers:executing-plans，代码批次遵循 superpowers:test-driven-development 与 verification-before-completion。本文为 #45 阶段二确认的批次设计；源码实施与验证按下列步骤及 Issue 记录推进。

**目标：** 全后端统一采用示例中的分层、根/模块 GroupApp、职责子 Router 和 request/response 公共组织。

**设计依据：** [迁移设计](../specs/2026-11-02-gva-backend-migration-design.md)、[后端规范](../../modules/backend-layer-rules.md)、[接口契约](../../frontend-backend/boundary.md)。

**技术栈：** 现有 Go/Gin/GORM，不更换依赖或部署配置；本批结构迁移不改前端、数据库模式或业务协议。

## 0. 影响面与基线（进入实现前）

- [x] 重新读取 #44/#45 状态和主干基线，决定 #44 与迁移的先后；不将未合并协议变更混入结构批次。
- [x] 通过 codebase-memory-mcp 补全每批旧/新符号、调用方、构造/后台任务/中间件及测试清单。当前已定位三个 Module、三个 Handlers、三个 Router、三个 dto.go，不能以定位替代完整字段及入站调用核查。
- [x] 记录全部真实端点方法/路径与 public/private/ProjectScope/协议标记清单，以及 JSON 字段、HTTP 状态、CSV/探活快照。
- [x] 在独立临时 MySQL 上运行 `cd server && go test -count=1 -p 1 -json ./...`，并运行 `go build ./...`，记录测试数、失败与跳过项；有业务变化的基线问题单独处理。

## 1. 公共类型与 Request/Response 归位

**文件边界：** `server/model/{system,project,cloud}/dto.go` → 各模块 `request/`、`response/`；`model/common/response.go`、`utils/response/` → `model/common/{request,response}/`；调用方涉及 API、Service/Store、middleware、initialize、Swagger 和测试。

- [x] 先写/补充外部 JSON、参数归一化和错误/CSV 快照测试，运行确认能检测不兼容类型或字段改变。
- [x] 逐模块搬迁输入与外部结果，保留内部查询结果的明确归属；统一 PageInfo/分页/响应公共类型。每次搬迁同步全部 import，不以长期 type alias 保留双入口。
- [x] 更新 API/Service 参数和返回类型、Swagger，保持默认20/上限200、id/时间字段及既有零值/null/空数组语义。
- [x] 运行受影响 package 定向测试，随后全量 `go test ./...`、`go build ./...`；响应相关另跑 `go test -race ./initialize ./middleware ./model/common/response`。实际目录随最终落位核对。
- [x] 独立提交；回退本批类型和全部调用方，不能仅回退类型定义。

## 2. 公共实体组合

**文件边界：** `server/global/model.go`（新增目标）；现有 `server/model/{system,project,cloud}/` 实体及 Store/Service/seed/测试中的实体构造。

- [x] 逐实体记录现有主键、时间字段、JSON/gorm 标签和删除行为；先补字段序列化/表结构元数据及删除回归。
- [x] 仅对字段集相同的实体提取本项目基础结构；同步因匿名嵌入改变而无法使用提升字段键的复合字面量。
- [x] 不引入 DeletedAt/新列，不更改物理删除或既有时间语义。不能保持行为的实体保留单独结构并记录原因，而非借模板自动改模式。
- [x] 定向与隔离 MySQL 全量回归、`go build ./...` 通过后独立提交；回退基础结构及全部实体构造。

## 3. API/Service GroupApp 与唯一生产装配链

**文件边界：** `server/service/enter.go`、各模块 `service/*/enter.go`；`api/v1/enter.go`、各模块 `api/v1/*/enter.go` 与 `handlers.go`；`initialize/router.go`、`initialize/system_module.go`、`initialize/{project,cloud}/module.go`；所有依赖这些入口的路由、中间件、后台任务及测试。

- [x] 先写 GroupApp 实际消费、启动初始化顺序、缺依赖503、真实路由注入替身及恢复测试；先失败再切换入口。
- [x] 根级具名聚合、模块级职责聚合，并把真正生产消费改到 ApiGroupApp/ServiceGroupApp；别名引用稳定对象，禁止复制装配前的有状态服务值。
- [x] 现有构造逻辑进入统一服务组装配，保留配置、密钥处理、SDK registry、RBAC、ProjectScope、审计清理及依赖失败行为；不顺带修正既有业务规则。
- [x] 同批删除对应 New*Module/New*Handlers 的平行生产入口，更新全部测试替换点；全局替换的测试串行隔离并 cleanup，不让测试污染运行态。
- [x] 全量隔离数据库测试、`go build ./...`、`go test -race ./initialize ./middleware ./router`，以及三种前缀的路由/权限/协议回归通过后独立提交；回退整个装配批次。

## 4. 全模块职责子 Router

**文件边界：** `server/router/system/`、`router/project/`、`router/cloud/` 的 `enter.go` 和各职责文件；`initialize/router.go` 注册调用及实际路由测试。

- [x] 基于原真实路由清单先写特征测试，覆盖方法/路径/参数/public-private、中间件拒绝、项目范围、默认/自定义/根前缀和探活/CSV。
- [x] system 拆 base/auth/user/role/ldap/audit；project 拆项目/成员；cloud 拆账号/资产。模块 RouterGroup 匿名嵌入并显式调用各 InitXxxRouter，API 别名从模块 enter.go 集中取得。
- [x] 保持注册顺序和 JWT 前协议标记；不重复绑定、不漏端点、不重新创建丢失权限的组。全局规范不能以“只迁 system”结案。
- [x] `go test -count=1 ./initialize ./router ./middleware`、相关 race、全量 `go test ./...` 和 `go build ./...`；运行启动/health/ping 及必要实际接口验证后提交。回退对应模块所有职责文件与聚合注册调用。

## 5. 迁移完成核对

- [x] 通过图谱核实所有模块均使用新入口，无遗留旧包、别名、兼容包装或重复生产链。
- [x] 隔离 MySQL 全量测试、构建及适用 race 通过，测试环境清理；记录未覆盖的外部消费者和真实环境验证。
- [x] 保持 #18 协议迁移与结构提交分开；若结构确实导致前端需要修改字段/协议，停止该部分并回到契约设计，不擅自扩展范围。
- [x] 更新规范的实际状态和 #45 批次提交、验证/回退记录；后端组织方式对照复审通过，用户已授权提交PR并合入main。#45其他范围仍独立跟踪，不自动关闭总Issue。

- [x] 纠正项验收：职责类型归各自文件，enter仅聚合Group；局部操作使用命名request DTO，分页查询内嵌PageInfo并显式映射Service条件。入口布局、请求JSON与查询映射守卫通过，用户AuthApi改动逐字保留；实际验证结果和适配差异见#45。
