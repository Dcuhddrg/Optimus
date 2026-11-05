# 前端示例

讲解型示例，按 Optimus 的 React + TypeScript 技术栈改写自 gin-vue-admin `cee04b08` 的 `aiDoc/examples/frontend`（原为 Vue/Pinia）。示例中的 `order` 模块是占位业务，不代表已实现；写法以每篇末尾的真实参考文件为准。

## 包含内容

| 示例 | 对应上游 | 说明 |
| --- | --- | --- |
| [api-example.md](api-example.md) | `api-example.md` | 模块接口函数 |
| [context-example.md](context-example.md) | `pinia-example.md` | 跨页面共享状态（Context + hook，替代 Pinia） |
| [view-example.md](view-example.md) | `view-example.md` | 页面组件 |
| [reuse-example.md](reuse-example.md) | `utils-usage-example.md` | 复用公共请求、组件与多语言能力 |

## 适用范围

新增或修改以下文件前先阅读对应示例：

- `web/src/api/**`
- `web/src/auth/**`（共享状态 Context）
- `web/src/view/**`
- `web/src/components/**`
- `web/src/plugin/**`
