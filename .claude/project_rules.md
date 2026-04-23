# CLAUDE.md

> 本文件是 Claude Code 的行为契约。每次会话开始时必须完整阅读。

---

## 🔴 BOUNDARIES — 绝对禁止（优先级最高）

在没有明确指令的情况下，Claude **绝对禁止**：

- **重构未涉及的文件** — 只修改任务明确指定的文件
- **更改技术选型** — 不替换已有的库、框架或工具
- **删除注释或 TODO** — 注释是团队知识，不可擅自删除
- **自动升级依赖版本** — package.json 的版本变更需人工确认
- **修改环境配置文件** — .env, .env.example 非必要不触碰
- **跨越模块边界** — 前端任务不改后端代码，反之亦然

**违反以上规则时，必须先暂停并明确提问。**

---

## 🏗️ CONTEXT — 项目架构

### 项目结构

\`\`\`
project-root/
├── frontend/          # React + TypeScript
│   ├── src/
│   │   ├── components/   # 纯展示组件（无业务逻辑）
│   │   ├── features/     # 业务模块（含状态、API调用）
│   │   ├── hooks/        # 自定义 hooks
│   │   └── lib/          # 工具函数、常量
├── backend/           # Node.js + Express / (或你的后端)
│   ├── src/
│   │   ├── routes/       # 路由层（仅参数校验）
│   │   ├── services/     # 业务逻辑层
│   │   ├── repositories/ # 数据访问层
│   │   └── models/       # 数据模型
└── shared/            # 前后端共享类型定义
\`\`\`

### 关键约定

- **前端状态管理**：[Zustand / Redux / Jotai — 填写你用的]
- **API 通信**：所有 API 调用统一走 `frontend/src/lib/api.ts`
- **数据库访问**：只能通过 `repositories/` 层，禁止在 `routes/` 直接查询DB
- **类型共享**：前后端共用类型定义在 `shared/types/`

### 当前技术栈

| 层 | 技术 |
|----|------|
| 前端框架 | React 18 + TypeScript |
| 样式 | Tailwind CSS |
| 后端 | Node.js + Express |
| 数据库 | PostgreSQL + Prisma |
| 测试 | Vitest + Playwright |

---

## 📐 CONVENTIONS — 代码规范

### TypeScript

\`\`\`typescript
// ✅ 正确：使用 interface 定义对象类型
interface UserProfile {
  id: string;
  email: string;
  createdAt: Date;
}

// ❌ 错误：不用 any，不用 type alias 替代 interface
const handler = (data: any) => { ... }
\`\`\`

### React 组件

\`\`\`typescript
// ✅ 正确：函数组件 + 明确 props 类型
interface ButtonProps {
  label: string;
  onClick: () => void;
  variant?: 'primary' | 'secondary';
}

export function Button({ label, onClick, variant = 'primary' }: ButtonProps) { ... }

// ❌ 错误：不使用 default export，不用 React.FC
export default function Button(props: any) { ... }
\`\`\`

### API 错误处理

\`\`\`typescript
// ✅ 统一使用自定义 AppError
throw new AppError('USER_NOT_FOUND', 404, '用户不存在');

// ❌ 不直接 throw Error 或返回裸字符串
throw new Error('not found');
\`\`\`

### 命名规则

| 类型 | 规则 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `UserCard.tsx` |
| 工具函数 | camelCase | `formatDate.ts` |
| 常量 | SCREAMING_SNAKE | `MAX_RETRY_COUNT` |
| CSS类名 | kebab-case | `user-card__title` |

### 注释语言

- **代码注释**：中文（团队母语）
- **JSDoc**：英文（工具链兼容性）
- **Git commit**：`类型(范围): 描述` — 使用中文描述

---

## 📤 FORMAT — 输出格式控制

### 回复语言

- **对话交流**：中文
- **代码本身**：英文变量名/函数名
- **错误分析**：中文解释 + 英文代码示例

### 代码变更格式

每次提交代码变更时，必须遵循以下结构：

\`\`\`
## 变更说明
[1-2句话说明做了什么，为什么这样做]

## 影响范围
- 修改文件：[列出所有变更的文件]
- 影响功能：[可能影响的功能点]
- 未修改：[明确说明哪些相关文件没有动]

## 代码
[代码块]

## 注意事项
[需要人工确认的点，如果有的话]
\`\`\`

### 禁止的输出习惯

- ❌ 不要说"我已经完成了X，还顺便优化了Y" — 只做被要求的事
- ❌ 不要在没有被问到的情况下提供"更好的方案"
- ❌ 不要生成超过任务范围的测试文件
- ✅ 如果发现潜在问题，**先完成任务**，然后在"注意事项"里提及

---

## 🔄 工作流程约定

### 接受任务时的标准流程

1. **复述任务** — 用一句话确认你理解的任务范围
2. **列出将修改的文件** — 明确边界
3. **如有歧义，先提问** — 不要假设
4. **执行** — 只做约定范围内的事
5. **汇报** — 按 FORMAT 格式说明变更

### 遇到以下情况必须暂停并提问

- 需要修改"边界"章节禁止的内容时
- 发现现有代码有严重 bug，但不在任务范围内
- 不确定应该改前端还是后端
- 任务描述与现有架构有冲突

---

## 📝 项目特定知识（持续更新）

### 已知技术债务（不要"修复"这些）
- `UserContext.tsx` 存在重渲染问题，正在计划重构，暂不修改
- `legacy/` 目录下的代码是历史遗留，不要碰

### 重要业务规则
- 用户 ID 永远是 UUID，不是数字自增
- 所有时间存储为 UTC，展示时转为用户时区
- 软删除：不物理删除记录，使用 `deletedAt` 字段