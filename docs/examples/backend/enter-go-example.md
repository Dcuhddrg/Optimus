# enter.go 示例

> 改写自 [gin-vue-admin / backend/enter-go-example.md](https://github.com/flipped-aurora/gin-vue-admin/blob/cee04b08d1262036ca33e3026e1e13d29471c7d5/aiDoc/examples/backend/enter-go-example.md)（`cee04b08`）。2026-10-30 适配 Optimus：保留上游讲解结构，代码、路径及现状说明已修改；上游许可见 [LICENSE.gin-vue-admin](../LICENSE.gin-vue-admin)。

## 这个文件负责什么

enter.go 汇总现有路由分组；initialize 构造模块并注入 API handlers。它不是要求每层必须新建一个全局 GroupApp。

## 什么时候应该这样写

- 给现有模块扩充路由。
- 新增模块需要纳入启动装配。
- 检查重复注册或循环依赖。

## 推荐写法示例

当前装配关系：

```text
main
  → initialize/router.go：创建模块与 handlers，划分公开/私有路由
  → router/enter.go：RouterGroupApp 汇总
  → router/system/enter.go：InitSystemRouter(..., handlers)
  → API handler → Service → Model / Store
```

根级 ServiceGroupApp 已按用户决定移除。上游 `service/enter.go`、`api/v1/enter.go` 的全局示例只解释其仓库模式，不是本仓库应恢复的入口。新模块需参考现有构造/注入链，同时注册路由与权限种子；不能仅加一个 enter.go 就视为接入完成。

## 为什么这样写

沿用真实组合入口，避免多套注册。分层职责由既有规范维护，本页说明文件间如何连接。

## 常见错误

- 复制上游 GroupApp 造成平行装配。
- handler 已构造但路由、权限或前端调用没有接入。
- 将已移除旧入口作为模板。

## 真实参考文件

- [根 RouterGroup](../../../server/router/enter.go)、[系统 RouterGroup](../../../server/router/system/enter.go)。
- [SystemModule](../../../server/initialize/system_module.go)、[路由装配](../../../server/initialize/router.go)、[装配测试](../../../server/initialize/system_module_test.go)。
