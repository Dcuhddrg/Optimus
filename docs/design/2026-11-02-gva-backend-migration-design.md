# 参考 gin-vue-admin 的后端结构迁移

关联 #45 阶段二。用户确认后端迁移“参考 gin-vue-admin 的”，目标方向已确定；本设计细化对应关系和兼容边界，不宣称源码已经迁移，也不替代后续批次验证和验收。参考版本为已核对示例及实际源文件的 `cee04b08d1262036ca33e3026e1e13d29471c7d5`。

## 目标与替代关系

所有后端模块及插件遵守 `Router → API → Service → Model`，根级采用上游的 RouterGroupApp、ApiGroupApp、ServiceGroupApp，模块级以 enter.go 聚合职责类型和集中别名。当前 system、project、cloud 均须迁移；先后分批不构成长期例外。

本设计替代后端规范中“永久保留逐模块 handlers 注入、禁止全局 API/Service 聚合”的旧适配结论。#11 对旧 usecase/infra 的清理和已确认外部行为继续保留。前端 React 技术栈、#18 的响应迁移及插件注册框架不在本次结构迁移中变更。

## 各层目标

| 位置 | 目标 | 迁移前定位与需要切换的消费者 |
| --- | --- | --- |
| server/service/enter.go、service/<module>/enter.go | 根级 ServiceGroupApp 具名聚合模块，模块聚合职责服务 | 当前 NewSystemModule、NewProjectModule、NewCloudModule 构造服务；API handler、JWT/RBAC、ProjectScope、审计保留任务等需改由同一组取得服务 |
| server/api/v1/enter.go、api/v1/<module>/enter.go | 根级 ApiGroupApp、模块 ApiGroup，集中 service 别名 | 当前 SystemHandlers/CloudHandlers/ProjectHandlers 与 New*Handlers；路由、initialize 及注入测试同步切换，职责 handler 可按 XxxApi 命名统一 |
| server/router/enter.go、router/<module>/enter.go | 根级模块聚合，模块匿名嵌入职责子 Router，集中 api 别名 | 当前三个 Init*Router 均集中端点；拆到职责文件并由 initialize 直接调用注册，保持全部方法、路径和中间件顺序 |
| model/<module>/request/ | 请求与查询 DTO，按职责分文件，通用分页复用 PageInfo | 当前三个 model/<module>/dto.go 混合请求与结果；API 绑定、Service 参数、测试及 Swagger import 同批更新 |
| model/<module>/response/ | 面向 HTTP 的结果 DTO，与实体及内部查询结果分离 | AuditListData、用户/项目/云资源 View/Result 移到响应职责；内部 Store/Service 结果按用途归类，不伪装成外部协议 |
| model/common/request/、model/common/response/ | PageInfo、统一响应与分页公共类型/助手 | 现有 utils/response、model/common/response.go 和模块分页需盘点后合并职责；路由协议标记、错误/CSV/Recovery 调用一起验证，不保留重复定义的兼容包 |
| global/model.go | 参考 GVA_MODEL 的公共基础模型组织方式，采用本项目名称与字段契约 | 逐实体核对 ID、时间、gorm 标签、复合主键和现有行为；不能机械让所有模型继承同一个类型 |

具体公共类型名称及每个职责类型的重命名在对应批次清单固定；所有替换必须有完整的旧符号→新符号→调用方→测试映射，不通过别名掩盖未完成迁移。

## 聚合与生命周期

生产链统一为：基础设施准备 → initialize 装配 ServiceGroupApp → ApiGroupApp 中的职责 API → 模块 RouterGroup 注册 → API → Service → Model/数据访问。

根级与模块级聚合对象在进程内保持稳定地址。构造函数用于创建模块服务和 LDAP、云客户端等外部依赖。查询和事务由职责 Service 使用启动时发布的 global.OPTIMUS_DB 或当前事务句柄执行。测试替身可以通过接口传入。

ServiceGroup 按参考示例匿名嵌入具体服务值；API 的 Service 别名和 Router 的 API 别名按示例复制聚合值。本项目认证、LDAP 和云服务含启动期依赖，包初始化时的值副本尚未装配，因此 initialize 在装配全部服务后、注册路由前同步 API 别名。请求期间不得刷新；测试替换和恢复服务组时同步恢复别名。不能仅增设未被实际消费的 GroupApp 外壳。

无数据库、缺服务和初始化失败时保留现有返回与启动行为；不能把原本的 503 变成 nil 指针 panic。RBAC capability、项目范围解析、审计清理及云同步依赖必须迁入同一服务生命周期。组全局状态的测试不得并行修改同一实例，替身装配后通过 HTTP 验证真实路径，并在 cleanup 中恢复字段；不得在请求期间重装全局组。

## 必要兼容边界

已核对上游 global/model.go：GVA_MODEL 使用 json ID、默认时间字段名及 DeletedAt；上游 PageInfo.Paginate 默认条数 10、最大 100。Optimus 现有契约为小写 id、camelCase 时间字段、分页默认20/上限200，且部分表没有软删除。这些差异按既有接口及数据行为保留：

- 公共模型可统一重复字段，但保留现有 Go/JSON/gorm 类型和列；不自动新增 DeletedAt、改变删除语义或触发表结构迁移。含复合键、只创建时间或其他不同字段集的实体使用符合真实模型的公共组合，不为统一外观硬套字段。
- PageInfo 复用结构与分页职责，保留既有参数来源、默认值、上下限和非法值处理；普通集合不伪造分页。绑定结构变更不把原来静默忽略的条件改成请求失败。
- 统一响应工具保持当前 HTTP 状态、code 0/7、普通无数据 {}、401 null；旧端点继续按 #18 已登记状态输出，不以结构搬迁执行全局协议切换。CSV、探活、已开始写出的响应边界保持。
- 不复制上游订单业务、OperationRecord、自动插件发现、菜单/字典初始化等未确认业务或运行机制。

这些是保持结构迁移外部行为的兼容要求，不将存量目录/装配方式固定为目标。需要改契约或数据语义时独立确认和实施。

## 分批顺序与验证

1. 公共请求/响应和模块 request/response 归位：先建立调用方与字段快照，纯搬迁保持行为；公共基础模型按实体分批验证，无模式迁移。
2. 统一 GroupApp 装配：根级、三个模块组、API/Service 集中别名、initialize、中间件能力、后台任务和测试同批切换，清除对应旧生产装配入口。
3. 三个模块子 Router：system 的 base/auth/user/role/ldap/audit；project 的项目/成员；cloud 的账号/资产。规范覆盖全局，可分提交但须保留全部端点与真实继承关系。
4. 全局清理与验收：无旧导入/平行入口，契约快照、全量测试/构建、适用 race、隔离数据库及启动/探活验证通过。

#44 仍是未合并的角色协议候选，改动与 system API/router 重叠；开始迁移批次时重新核实其状态，明确选择先合并后迁移或之后重放协议变更，不能暗中包含其未合并功能。详细执行步骤见分批计划。

本文记录阶段二确认的目标和迁移前对应关系，实际批次实现与验证见计划和 #45。每批执行前补全影响面；未核查清单、未实施代码和用户验收不得据目标设计勾选为已完成。
