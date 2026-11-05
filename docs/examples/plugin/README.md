# 插件示例

## 包含内容

[完整插件示例](full-plugin-example.md) 按 gin-vue-admin `cee04b08` 改写，说明前后端目录和实现顺序；入口写法见 [plugin.go 示例](../backend/plugin-go-example.md)。注册方式见 [ADR-004](../../adr/0004-gva-plugin-registration.md)，框架代码随第一个插件落地，当前仓库中不存在。

## 适用范围

仅用于可整体启用或移除、且不属于 system/project/cloud 的能力，见[插件开发约束](../../modules/plugin-development.md#什么时候做成插件)；其他功能按普通模块开发。
