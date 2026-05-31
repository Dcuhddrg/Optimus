# Service Card 尺寸统一化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 消除服务门户首页卡片网格的「点击框大小不一」问题：删除 `hot/.cardWide` 跨列差异化，所有卡片同尺寸；「热门」改用 `servicePortal.tag.popular` tag + 右上角角标 + tag-based 过滤表达。

**Architecture:** 改造分 4 个独立 commit：i18n key 先行 → 类型/数据迁移 → 组件渲染 → 样式收尾。每步通过 `pnpm exec tsc -b` 才进下一步。无单元测试（项目无前端测试框架），收尾靠类型检查 + 视觉自检。

**Tech Stack:** React 18.3 + TypeScript 5.6 + CSS Modules + pnpm 10 + Vite 5.4。i18n 通过 `web/src/locale/messages.ts` 的 `MessageKey = typeof en` 类型推导 + `zhCN: typeof en` 约束。

**Spec:** `docs/superpowers/specs/2026-05-31-service-card-sizing-design.md`

**Pre-flight:**
- 工作目录：`/home/lilin/trae-project/Optimus`
- Web 子目录：`web/`，包管理器：`pnpm`（通过 `npx pnpm@10` 调用，corepack 在此环境不可用）
- `pnpm-lock.yaml` 已存在；安装产物在 `web/node_modules/`
- 没有前端测试命令；唯一校验手段：`cd web && npx pnpm@10 exec tsc -b`

---

## Task 1: 新增 `servicePortal.tag.popular` i18n key

**Files:**
- Modify: `web/src/locale/messages.ts:373-382` (英文 tag 块)
- Modify: `web/src/locale/messages.ts:804-813` (中文 tag 块)

- [ ] **Step 1: 在英文字典添加 popular**

打开 `web/src/locale/messages.ts`，找到 `en.servicePortal.tag` 段（约第 373 行）。当前内容：

```ts
    tag: {
      approval: 'Approval',
      oneToTwoDays: '1-2 days',
      projectScope: 'Project scoped',
      costReview: 'Cost review',
      securityReview: 'Security review',
      frequent: 'Frequent',
      fastResponse: 'Fast response',
      trackable: 'Trackable',
    },
```

改为：

```ts
    tag: {
      approval: 'Approval',
      oneToTwoDays: '1-2 days',
      projectScope: 'Project scoped',
      costReview: 'Cost review',
      securityReview: 'Security review',
      frequent: 'Frequent',
      fastResponse: 'Fast response',
      trackable: 'Trackable',
      popular: 'Popular',
    },
```

- [ ] **Step 2: 在中文字典添加 popular**

同文件，找到 `zhCN.servicePortal.tag` 段（约第 804 行）。当前内容：

```ts
    tag: {
      approval: '需审批',
      oneToTwoDays: '1-2 天',
      projectScope: '项目范围',
      costReview: '成本确认',
      securityReview: '安全审核',
      frequent: '高频',
      fastResponse: '快速响应',
      trackable: '可追踪',
    },
```

改为：

```ts
    tag: {
      approval: '需审批',
      oneToTwoDays: '1-2 天',
      projectScope: '项目范围',
      costReview: '成本确认',
      securityReview: '安全审核',
      frequent: '高频',
      fastResponse: '快速响应',
      trackable: '可追踪',
      popular: '热门',
    },
```

- [ ] **Step 3: typecheck**

Run: `cd /home/lilin/trae-project/Optimus/web && npx pnpm@10 exec tsc -b`
Expected: 无输出（即成功）。

如果 `zhCN` 段没加 popular 但 `en` 加了，TS 会报：`Property 'popular' is missing in type ... but required in type ...` — 这是 `zhCN: typeof en` 约束的预期保护。

- [ ] **Step 4: 提交**

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
feat(service-portal): 新增 servicePortal.tag.popular i18n key

为后续卡片尺寸统一化（hot → tag 改造）做准备：en/zhCN 字典同步新增 popular 词条。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: 类型 + catalog 数据迁移

**Files:**
- Modify: `web/src/view/service/ticketTypes.ts` (删除 `ServiceCatalogItem.hot` 字段)
- Modify: `web/src/view/service/data/serviceCatalog.ts` (4 处删除 `hot: true` + 4 处给 `tags` 追加 popular + 改 `listByCategory`)

- [ ] **Step 1: 从 ServiceCatalogItem 类型删除 hot**

打开 `web/src/view/service/ticketTypes.ts`，查找包含 `hot?` 的行。例如：

```ts
export interface ServiceCatalogItem {
  id: string;
  category: ServiceCategory;
  i18nKey: string;
  iconToken: 'shield' | 'cloud' | 'network' | 'env' | 'support';
  tags: string[];
  hot?: boolean;
  estimatedSla: string;
}
```

删除 `hot?: boolean;` 一行：

```ts
export interface ServiceCatalogItem {
  id: string;
  category: ServiceCategory;
  i18nKey: string;
  iconToken: 'shield' | 'cloud' | 'network' | 'env' | 'support';
  tags: string[];
  estimatedSla: string;
}
```

- [ ] **Step 2: 在 serviceCatalog.ts 顶部声明 POPULAR_TAG 常量**

打开 `web/src/view/service/data/serviceCatalog.ts`。在 `SERVICE_CATALOG` 数组定义之前（约第 14 行 import 后）加入：

```ts
const POPULAR_TAG = 'servicePortal.tag.popular';
```

- [ ] **Step 3: 给 4 条 hot 项追加 popular tag、删除 hot 字段**

`SERVICE_CATALOG` 数组中 4 条 `hot: true` 项分别是 `account.role-grant` / `cloud.compute-resize` / `network.vpn-access` / `support.incident`。

对每一条执行同样的改造：把 `hot: true` 那一行删除，同时往 `tags` 数组里追加 `POPULAR_TAG`。

`account.role-grant`：
```ts
  {
    id: 'account.role-grant',
    category: 'account',
    i18nKey: 'servicePortal.catalog.accountRoleGrant',
    iconToken: 'shield',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.oneToTwoDays'],
    hot: true,
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
```
改为：
```ts
  {
    id: 'account.role-grant',
    category: 'account',
    i18nKey: 'servicePortal.catalog.accountRoleGrant',
    iconToken: 'shield',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.oneToTwoDays', POPULAR_TAG],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
```

`cloud.compute-resize`：
```ts
    tags: ['servicePortal.tag.projectScope', 'servicePortal.tag.costReview'],
    hot: true,
```
改为：
```ts
    tags: ['servicePortal.tag.projectScope', 'servicePortal.tag.costReview', POPULAR_TAG],
```

`network.vpn-access`：
```ts
    tags: ['servicePortal.tag.securityReview', 'servicePortal.tag.frequent'],
    hot: true,
```
改为：
```ts
    tags: ['servicePortal.tag.securityReview', 'servicePortal.tag.frequent', POPULAR_TAG],
```

`support.incident`：
```ts
    tags: ['servicePortal.tag.fastResponse', 'servicePortal.tag.trackable'],
    hot: true,
```
改为：
```ts
    tags: ['servicePortal.tag.fastResponse', 'servicePortal.tag.trackable', POPULAR_TAG],
```

- [ ] **Step 4: 改写 listByCategory**

同文件，找到 `listByCategory` 函数（约第 131 行）。当前实现：

```ts
export function listByCategory(category: ServiceCategory | 'popular'): ServiceCatalogItem[] {
  if (category === 'popular') {
    return SERVICE_CATALOG.filter((item) => item.hot);
  }
  return SERVICE_CATALOG.filter((item) => item.category === category);
}
```

改为：

```ts
export function listByCategory(category: ServiceCategory | 'popular'): ServiceCatalogItem[] {
  if (category === 'popular') {
    return SERVICE_CATALOG.filter((item) => item.tags.includes(POPULAR_TAG));
  }
  return SERVICE_CATALOG.filter((item) => item.category === category);
}
```

- [ ] **Step 5: typecheck**

Run: `cd /home/lilin/trae-project/Optimus/web && npx pnpm@10 exec tsc -b`
Expected: 无输出。

可能出现的错误：
- 如果 `ServiceCard.tsx` 仍引用 `item.hot`，会报 `Property 'hot' does not exist on type 'ServiceCatalogItem'`。这是预期的 — Task 3 会修复。
  - 解决：本 task 暂不处理；直接进 Task 3，typecheck 会在 Task 3 步骤 4 再次通过。

如果上面这个错误是唯一的报错，把它视为预期的、可继续进入 Task 3 的信号。

- [ ] **Step 6: 提交**

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/view/service/ticketTypes.ts web/src/view/service/data/serviceCatalog.ts
git commit -m "$(cat <<'EOF'
refactor(service-portal): hot 字段迁移至 tags.popular

- ServiceCatalogItem.hot 字段移除
- 4 条原 hot:true 项目（role-grant/compute-resize/vpn-access/support-incident）的 tags 追加 servicePortal.tag.popular
- listByCategory('popular') 改为 tag-based 过滤

ServiceCard.tsx 暂时残留 item.hot 引用，typecheck 在下一个 commit 通过。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: ServiceCard 组件改造（删 cardWide + 加 hotBadge）

**Files:**
- Modify: `web/src/view/service/components/ServiceCard.tsx`

- [ ] **Step 1: 重写 ServiceCard.tsx**

打开 `web/src/view/service/components/ServiceCard.tsx`。当前文件（C1 之后的状态）：

```tsx
// ServiceCard.tsx
import { ShieldCheck, Cloud, ShareNetwork, StackSimple, Headset } from '@phosphor-icons/react';

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { ServiceCatalogItem } from '../ticketTypes';
import styles from './ServiceCard.module.css';

interface Props {
  item: ServiceCatalogItem;
  onSelect: (item: ServiceCatalogItem) => void;
}

export function ServiceCard({ item, onSelect }: Props) {
  const { t } = useI18n();
  const titleKey = `${item.i18nKey}.title` as MessageKey;
  const descKey = `${item.i18nKey}.description` as MessageKey;
  return (
    <button type="button" className={item.hot ? `${styles.card} ${styles.cardWide}` : styles.card} onClick={() => onSelect(item)}>
      <span className={styles.icon} aria-hidden="true">
        <ServiceTokenIcon token={item.iconToken} />
      </span>
      <h3 className={styles.title}>{t(titleKey)}</h3>
      <p className={styles.description}>{t(descKey)}</p>
      <div className={styles.tags}>
        {item.tags.map((tag) => (
          <span key={tag} className={styles.tag}>
            {t(tag as MessageKey)}
          </span>
        ))}
      </div>
    </button>
  );
}

function ServiceTokenIcon({ token }: { token: ServiceCatalogItem['iconToken'] }) {
  // ...
}
```

整体替换为：

```tsx
// ServiceCard.tsx
import { ShieldCheck, Cloud, ShareNetwork, StackSimple, Headset } from '@phosphor-icons/react';

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { ServiceCatalogItem } from '../ticketTypes';
import styles from './ServiceCard.module.css';

const POPULAR_TAG = 'servicePortal.tag.popular';

interface Props {
  item: ServiceCatalogItem;
  onSelect: (item: ServiceCatalogItem) => void;
}

export function ServiceCard({ item, onSelect }: Props) {
  const { t } = useI18n();
  const titleKey = `${item.i18nKey}.title` as MessageKey;
  const descKey = `${item.i18nKey}.description` as MessageKey;
  const isPopular = item.tags.includes(POPULAR_TAG);
  return (
    <button type="button" className={styles.card} onClick={() => onSelect(item)}>
      {isPopular ? (
        <span className={styles.hotBadge} aria-hidden="true">
          {t('servicePortal.tag.popular')}
        </span>
      ) : null}
      <span className={styles.icon} aria-hidden="true">
        <ServiceTokenIcon token={item.iconToken} />
      </span>
      <h3 className={styles.title}>{t(titleKey)}</h3>
      <p className={styles.description}>{t(descKey)}</p>
      <div className={styles.tags}>
        {item.tags.map((tag) => (
          <span key={tag} className={styles.tag}>
            {t(tag as MessageKey)}
          </span>
        ))}
      </div>
    </button>
  );
}

function ServiceTokenIcon({ token }: { token: ServiceCatalogItem['iconToken'] }) {
  const size = 20;
  const weight = 'duotone' as const;
  switch (token) {
    case 'shield':
      return <ShieldCheck size={size} weight={weight} />;
    case 'cloud':
      return <Cloud size={size} weight={weight} />;
    case 'network':
      return <ShareNetwork size={size} weight={weight} />;
    case 'env':
      return <StackSimple size={size} weight={weight} />;
    case 'support':
      return <Headset size={size} weight={weight} />;
    default:
      return null;
  }
}
```

变更点：
- 顶部新增 `const POPULAR_TAG = 'servicePortal.tag.popular';`
- `className` 去除 `item.hot` 三元，仅 `styles.card`
- 新增 `isPopular` 判定 + 条件渲染 `<span className={styles.hotBadge}>`
- `ServiceTokenIcon` 函数体保持不变

- [ ] **Step 2: typecheck**

Run: `cd /home/lilin/trae-project/Optimus/web && npx pnpm@10 exec tsc -b`
Expected: 无输出。

如果报 `Property 'hotBadge' does not exist on type ...`：因为 CSS module 还没加 `.hotBadge` 规则，CSS Modules 的类型定义找不到。这是 Task 4 才会补上的规则；本步骤可暂时忽略此一项（CSS Modules 类型生成在 build 时；`tsc` 通过 `*.module.css.d.ts` 推断时，未声明的类名 → `any`，不会报错）。如果报错则说明项目配置严格，需要回到 Step 1 把 `styles.hotBadge` 改成 `(styles as any).hotBadge`，并在 Task 4 完成后改回。

实际项目中无 CSS Modules 类型生成（无 `.d.ts` 自动生成插件），`styles` 推导为 `Record<string, string>`，任何 `styles.foo` 都返回 `string`，不会报错。预计直接通过。

- [ ] **Step 3: 提交**

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/view/service/components/ServiceCard.tsx
git commit -m "$(cat <<'EOF'
refactor(service-portal): ServiceCard 删 cardWide 三元、新增 hotBadge 角标

- 单一 .card className，不再依赖 item.hot
- popular tag 通过右上角 hotBadge 表达
- ServiceTokenIcon 保持不变

CSS 规则在下一个 commit 补齐。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: CSS 收尾（卡片样式 + 网格布局）

**Files:**
- Modify: `web/src/view/service/components/ServiceCard.module.css` (删 `.cardWide`、加 `position: relative` 给 `.card`、新增 `.hotBadge`)
- Modify: `web/src/view/service/ServicePortalPage.module.css` (`.serviceGrid` 改为 auto-fill、去掉 `grid-auto-flow: dense`)

- [ ] **Step 1: 修 ServiceCard.module.css**

打开 `web/src/view/service/components/ServiceCard.module.css`。

(1) 第一处改动：给 `.card` 加 `position: relative`（让 hotBadge 绝对定位有锚点）。

当前：
```css
.card {
  display: grid;
  align-content: start;
  gap: 0.65rem;
  min-height: 11rem;
  padding: 0.95rem;
  border: 1px solid rgba(86, 70, 55, 0.1);
  border-radius: 1rem;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.66), rgba(246, 238, 228, 0.44));
  text-align: left;
  font: inherit;
  cursor: pointer;
  transition: transform 120ms ease, box-shadow 120ms ease;
}
```

改为（在 `display: grid;` 之前加 `position: relative;`）：

```css
.card {
  position: relative;
  display: grid;
  align-content: start;
  gap: 0.65rem;
  min-height: 11rem;
  padding: 0.95rem;
  border: 1px solid rgba(86, 70, 55, 0.1);
  border-radius: 1rem;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.66), rgba(246, 238, 228, 0.44));
  text-align: left;
  font: inherit;
  cursor: pointer;
  transition: transform 120ms ease, box-shadow 120ms ease;
}
```

(2) 第二处：删除 `.cardWide` 规则。

当前：
```css
.cardWide {
  grid-column: span 2;
}
```

删除整段（包含上一行空行 + 这 3 行）。

(3) 第三处：新增 `.hotBadge` 规则。把下面这段附加到文件末尾：

```css
.hotBadge {
  position: absolute;
  top: 0.55rem;
  right: 0.55rem;
  min-height: 1.05rem;
  padding: 0 0.4rem;
  border: 1px solid rgba(86, 70, 55, 0.18);
  border-radius: 999px;
  background: #fffaf2;
  color: #935d2d;
  font-size: 0.62rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  pointer-events: none;
}
```

- [ ] **Step 2: 修 ServicePortalPage.module.css**

打开 `web/src/view/service/ServicePortalPage.module.css`，找到 `.serviceGrid`（约第 97 行）。

预期当前形态（B 节 bento 改造后）类似：

```css
.serviceGrid {
  display: grid;
  grid-template-columns: repeat(...);
  grid-auto-flow: dense;
  gap: ...;
}
```

整段替换为：

```css
.serviceGrid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 0.85rem;
}
```

关键变更：
- `grid-template-columns` 改为 `repeat(auto-fill, minmax(220px, 1fr))`
- 删除 `grid-auto-flow: dense` 一行
- `gap` 统一为 `0.85rem`（如果当前不同也以这值为准）

- [ ] **Step 3: typecheck**

Run: `cd /home/lilin/trae-project/Optimus/web && npx pnpm@10 exec tsc -b`
Expected: 无输出。

- [ ] **Step 4: 视觉自检（启动 dev server）**

Run: `cd /home/lilin/trae-project/Optimus/web && npx pnpm@10 run dev` （后台启动）

打开浏览器 `http://localhost:5173/service-portal`（端口若被占用 Vite 会自动切到 5174/5175），逐项确认 spec「验证清单」：

- [ ] 切到 popular 分类，能看到 4 张带角标的卡（role-grant / compute-resize / vpn-access / support-incident）
- [ ] 切到 account / cloud / network / env / support 分类，卡片正常显示
- [ ] 用浏览器 DevTools 选中任意两张卡，对比 `getBoundingClientRect()` 的 width — 在同一列宽下应当一致
- [ ] 角标的中文「热门」 / 英文 "Popular" 显示正常（切换 locale 验证）
- [ ] Tab 键聚焦卡片，焦点框 outline 大小一致
- [ ] 鼠标 hover 任意一张带角标的卡，角标不消失、不变形

如所有项目通过 → Step 5。如某项失败 → 回到 Step 1/2 调整对应 CSS。

如果 dev server 因后端代理 404 等问题无法启动或访问受限（参考前述会话中类似问题），允许跳过视觉自检，仅以 typecheck + 静态阅读 diff 作为收尾依据，并在 commit message 里写明「视觉自检受 dev server 问题阻塞」。

- [ ] **Step 5: 提交**

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/view/service/components/ServiceCard.module.css web/src/view/service/ServicePortalPage.module.css
git commit -m "$(cat <<'EOF'
style(service-portal): 服务卡尺寸统一 + hotBadge 样式

- ServiceCard.module.css 删 .cardWide，.card 增 position:relative，新增 .hotBadge 角标
- ServicePortalPage.module.css serviceGrid 改 auto-fill / minmax(220px, 1fr)，去掉 grid-auto-flow:dense

收尾：所有点击框同尺寸，热门通过角标表达。

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>
EOF
)"
```

---

## Self-Review

**Spec coverage:**

| Spec 要求 | 实施任务 |
|---|---|
| i18n key `servicePortal.tag.popular` 同步加 en/zhCN | Task 1 |
| 删除 `ServiceCatalogItem.hot` 字段 | Task 2 Step 1 |
| 4 条原 hot 项 tags 追加 popular | Task 2 Step 3 |
| `listByCategory('popular')` 改为 tag 过滤 | Task 2 Step 4 |
| ServiceCard 删 `.cardWide` 三元 | Task 3 |
| ServiceCard 新增 hotBadge 渲染 | Task 3 |
| `.cardWide` CSS 删除 | Task 4 Step 1(2) |
| `.hotBadge` CSS 新增 | Task 4 Step 1(3) |
| `.card` 加 `position: relative` | Task 4 Step 1(1) |
| `.serviceGrid` 改 auto-fill / minmax(220px, 1fr) | Task 4 Step 2 |
| 删除 `grid-auto-flow: dense` | Task 4 Step 2 |
| 验证清单（视觉自检） | Task 4 Step 4 |

所有 spec 要求都有对应任务。

**Placeholder scan:** 无 TBD / TODO / "add appropriate error handling" 等占位。所有代码块完整。

**Type consistency:**
- `POPULAR_TAG` 常量值 `'servicePortal.tag.popular'` 在 Task 2 Step 2、Task 3 Step 1 出现 — 字面值一致。
- `isPopular` 判定逻辑 `item.tags.includes(POPULAR_TAG)` 在 Task 3 与 `listByCategory` 实现一致（Task 2 Step 4）。
- CSS 类名 `.hotBadge` 在 Task 3（TSX 引用 `styles.hotBadge`）与 Task 4（CSS 定义 `.hotBadge`）一致。
- Spec 文件命名 `servicePortal.tag.popular` vs 实施 i18n key 路径 `en.servicePortal.tag.popular` — 一致。

无类型/命名漂移。

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-31-service-card-sizing.md`. Two execution options:

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
