# 状态管理示例

## 这个文件负责什么

跨页面身份、项目或语言状态由现有 Context 提供；仅用于单个页面的筛选、分页、详情和 loading 保留在页面。

## 什么时候应该这样写

- 多页面需要共享身份或项目上下文。
- 需要统一副作用与取消。
- 需要分清未验证会话和已验证用户信息。

## 推荐写法示例

共享身份状态使用现有 AuthContext。真实消费 hook 摘录（AuthContext、useContext 来自所在文件）：

```tsx
export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error('useAuth must be used inside AuthProvider');
  }
  return value;
}
```

已挂载 AuthProvider 的组件或 hook 内的消费侧教学片段：

```tsx
const { current, loading, sessionError, retrySession } = useAuth();
```

Provider 挂载顺序与副作用以实际 router/main 为准。`RequireAuth` 在会话恢复中或失败时阻止受保护页面；非认证故障保留凭据并重试，不能只看本地 token。React 状态适配不要求新建 Pinia store。

## 为什么这样写

共享副作用集中，页面保留交互状态；卸载/新请求取消旧恢复，旧响应不能清理新凭据。不要为每个列表引入全局状态。

## 常见错误

- 把上游 Pinia/Vue 示例直接放进 React 项目。
- 在 Provider 未挂载时使用 hook。
- 认为 token 存在就代表会话已验证。

## 真实参考文件

- [AuthContext](../../../web/src/auth/AuthContext.tsx)、[守卫](../../../web/src/auth/RequireAuth.tsx)、[真实 Provider 回归](../../../web/src/auth/AuthContext.test.tsx)。
- [语言 Provider](../../../web/src/locale/LanguageProvider.tsx)、[项目 Provider](../../../web/src/auth/ProjectContext.tsx)。
