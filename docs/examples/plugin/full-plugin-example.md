# 完整插件示例（目录与装配边界）

## 这个文件负责什么

说明前后端文件职责和实施顺序。本页是设计参照，不宣称 Optimus 已有完整插件注册框架或已验证业务样板。

## 推荐目录结构

下面是按职责划分的候选布局，`<name>` 是占位名，不代表已存在的文件；真正创建前先确认设计与装配入口：

```text
server/plugin/<name>/
├── api/             HTTP handlers
├── initialize/      显式依赖构造、路由与必要初始化
├── model/           实体与 DTO
├── router/          路由分组；按确认后的 enter.go 组合方式组织
├── service/         业务及持久化/协议协作
└── plugin.go        入口协调；不假定存在自动注册 SDK

web/src/plugin/<name>/
├── api/             复用公共认证请求的模块函数
├── components/      仅插件内复用的组件
└── view/            React/TypeScript 页面
```

上游 config/form/request 等目录按实际需要增加，不为对称而创建空目录；前端 `.vue` 和 Pinia 改为本项目 React 模式。

## 推荐实现顺序

1. 确认普通模块还是插件，列清外部依赖、权限、项目范围与生命周期。
2. 定义 Model/DTO 与 Service，选择现有可验证实现作参照。
3. 编写 API/Router，并在已有 initialize/enter.go 链中明确接入位置。
4. 仅补实际需要的数据库、配置、权限或菜单初始化，不能假定上游初始化函数存在。
5. 接入前端 API、React 页面、路由、导航、权限及中英文。
6. 验证正常/失败/权限不足/空状态，记录运行命令和未覆盖范围。

## 为什么这样写

按职责组织，入口保持薄；构造、注册接口和自动加载若需要新增，先形成明确设计，不在示例中暗中引入。

## 常见错误

- 把目录树当作已实现框架或已验证代码。
- 复制 `plugin/v2`、interfaces.Register 或 GVA 的菜单/字典助手。
- 为新增插件重新建立全局 ServiceGroupApp 或绕开现有权限。

## 真实参考文件

- Optimus 当前可参照 [模块装配](../../../server/initialize/system_module.go)、[启动路由](../../../server/initialize/router.go)、[系统路由](../../../server/router/system/enter.go)，但它们不是完整插件实现。
- 本项目完整插件样板的存在与运行尚未核实，不提供伪造的本地 SDK/业务文件链接；现行目录边界见 [根规则](../../../AGENT.MD#插件与模块)。
