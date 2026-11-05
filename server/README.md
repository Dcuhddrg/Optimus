## 上游参考结构（gin-vue-admin `cee04b08` README 原文）

以下为上游 README 原文，仅作参考；其中 `packfile` 已不存在于上游源码，`model/request|response` 实际为 `model/<模块>/request|response`。Optimus 实际结构见下一节。

```shell
├── api
│   └── v1
├── config
├── core
├── docs
├── global
├── initialize
│   └── internal
├── middleware
├── model
│   ├── request
│   └── response
├── packfile
├── resource
│   ├── excel
│   ├── page
│   └── template
├── router
├── service
├── source
└── utils
    ├── timer
    └── upload
```

| 文件夹       | 说明                    | 描述                        |
| ------------ | ----------------------- | --------------------------- |
| `api`        | api层                   | api层 |
| `--v1`       | v1版本接口              | v1版本接口                  |
| `config`     | 配置包                  | config.yaml对应的配置结构体 |
| `core`       | 核心文件                | 核心组件(zap, viper, server)的初始化 |
| `docs`       | swagger文档目录         | swagger文档目录 |
| `global`     | 全局对象                | 全局对象 |
| `initialize` | 初始化 | router,redis,gorm,validator, timer的初始化 |
| `--internal` | 初始化内部函数 | gorm 的 longger 自定义,在此文件夹的函数只能由 `initialize` 层进行调用 |
| `middleware` | 中间件层 | 用于存放 `gin` 中间件代码 |
| `model`      | 模型层                  | 模型对应数据表              |
| `--request`  | 入参结构体              | 接收前端发送到后端的数据。  |
| `--response` | 出参结构体              | 返回给前端的数据结构体      |
| `packfile`   | 静态文件打包            | 静态文件打包 |
| `resource`   | 静态资源文件夹          | 负责存放静态文件                |
| `--excel` | excel导入导出默认路径 | excel导入导出默认路径 |
| `--page` | 表单生成器 | 表单生成器 打包后的dist |
| `--template` | 模板 | 模板文件夹,存放的是代码生成器的模板 |
| `router`     | 路由层                  | 路由层 |
| `service`    | service层               | 存放业务逻辑问题 |
| `source` | source层 | 存放初始化数据的函数 |
| `utils`      | 工具包                  | 工具函数封装            |
| `--timer` | timer | 定时器接口封装 |
| `--upload`      | oss                  | oss接口封装        |

## Optimus 当前结构

逐目录核对记录见 [#53](https://github.com/Dcuhddrg/Optimus/issues/53)。新增代码按下表放置，不为对齐目录树建立空目录。

| 文件夹 | 说明 | 与上游 |
| --- | --- | --- |
| `api/v1/{system,project,cloud}` | API 层，`enter.go` 聚合 ApiGroup 并在包初始化时复制服务别名 | 一致 |
| `config` | 配置结构体（JWT 时长为 `7d` 形式字符串） | 一致 |
| `core` | viper（`-c` → `OPTIMUS_CONFIG` → gin 模式文件 → `config.yaml`，无热更新）、zap、server | 一致；保留 `OPTIMUS_*` 环境覆盖 |
| `global` | 全局对象：配置、DB、日志、`BlackCache`、`OPTIMUS_Timer` 等 | 一致 |
| `initialize` | router、gorm、timer、cloud provider 注册、启动数据初始化 | 一致；无 Redis |
| `initialize/internal` | GORM 配置与日志 writer（`log-zap` 时写入 zap），仅供 `initialize` 调用 | 一致 |
| `middleware` | JWT（`x-token`、刷新、黑名单）、Casbin、ProjectScope、日志、Recovery | 一致；ProjectScope 为 Optimus 项目隔离 |
| `model/{system,project,cloud,common}` | 数据模型，`request`/`response` 位于各模块下 | 一致 |
| `router/{system,project,cloud}` | 职责子组路由 | 一致 |
| `service/{system,project,cloud}` | 无状态业务服务，依赖从 `global` 或 `utils` 即时获取 | 一致 |
| `source/{system,cloud,project}` | 种子数据初始化器，`RegisterInit` 排序，启动时运行 | 结构一致；每次启动迁移并幂等写入（无网页初始化） |
| `task` | 定时任务实现（`ClearTable` 清理过期审计日志与黑名单） | 一致 |
| `utils` | 工具函数（jwt、hash、claims、casbin、router prefix 等） | 一致 |
| `utils/timer` | robfig/cron 定时任务封装 | 一致 |
| `utils/ldap` | LDAP 客户端，按调用构建 | Optimus 特有（同 `utils/upload` 写法） |
| `docs` | Swagger | 暂缓 |
| `resource`、`utils/upload`、`packfile` | 代码生成模板、OSS 上传、静态打包 | 不适用 |
