# 前端开发与复用规范

与[接口契约](boundary.md)共同使用；项目规则以 [AGENT.MD](../../AGENT.MD) 为准。

## 请求与复用

- 页面和组件负责交互、展示；请求放在 `web/src/api/` 模块函数，公共 HTTP 行为集中在 [utils/request.ts](../../web/src/utils/request.ts)。插件 API 按插件边界组织，复用公共请求能力；组件不自行拼接认证或绕过服务端校验。
- 优先检查 `web/src/utils/`（`request`、`format`）、`web/src/auth/`、`web/src/locale/`、`web/src/components/` 和相关模块已有能力（见[工具函数约束](frontend-utils.md)与[复用示例](../examples/frontend/reuse-example.md)），不凭上游文档假定工具可用。确有重复需求时再提取通用能力，不为单次使用新增抽象或依赖。

## 命名与状态

- React 组件使用 PascalCase，变量和函数使用 camelCase；文件、状态和路由沿用所在模块的约定，不批量重命名存量文件。
- 共享状态优先沿用项目已有 Context；仅组件内使用的状态放在组件或相关 hook。前端类型与真实响应一致，不用类型断言掩盖协议差异。
- 导出接口函数和复杂交互在类型无法表达意图时补充必要说明。

## 样式与页面

- 沿用现有 CSS、组件和主题变量，不照搬上游 UnoCSS、Element Plus 或主题类名。避免无必要的内联样式，动态值可局部使用。
- 路由和页面变更考虑认证与已有权限入口；新增页面覆盖加载、空数据、失败和权限不足状态。
- 性能优化依据实际列表规模、网络和渲染瓶颈选择分页、懒加载等方案，不假定能力已存在。
