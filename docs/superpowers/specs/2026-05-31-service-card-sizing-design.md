# Service Portal · 卡片尺寸统一化设计

> 日期：2026-05-31
> 状态：Draft，待实现
> 上游依赖：[`2026-05-28-service-portal-design.md`](2026-05-28-service-portal-design.md)（服务门户首版设计），B 节 Bento Editorial 改造（commit `be78bc4`）

## 背景

服务门户首页（`/service-portal`）的服务卡片网格当前存在「点击框大小不一」的问题：4 条 `hot: true` 的目录项被渲染为跨 2 列的宽卡（`.cardWide`），其余 9 条占 1 列，配合 `grid-auto-flow: dense` 算法产生 Bento 拼贴效果。

视觉上有层级感，但作为「可点击的卡片」来讲：

- 点击区域尺寸不可预期，影响用户对「这是一个统一类型的入口」的认知
- 触控/键盘聚焦时，焦点框轮廓大小跳变
- 视口宽度变化时，dense 重排导致顺序不稳定

本次改造目标：**所有服务卡统一尺寸**，「热门」通过非尺寸方式表达。

## 范围

只修改 `web/src/view/service/` 下的 5 个文件 + `ticketTypes.ts`，不动以下已落定的部分：

- C1/C2/C3（Phosphor 图标 / framer-motion 动画 / Geist 字体）
- B 节 TicketTimeline 描线动画
- MyTicketsAside 空状态插画
- RequestDrawer 抽屉行为
- 工单列表 / 详情页 / 表单 schema

## 设计

### 卡片视觉合同

#### 网格
```css
.serviceGrid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 0.85rem;
}
```

- 移除 `grid-auto-flow: dense`
- 所有卡同尺寸、按视口宽度自然换行
- 卡片最小宽 220px，最大充满列

#### 卡片本体
- 单一 `.card` 类（删除 `.cardWide`）
- 不再读取 `item.hot` 字段
- 点击区域 = 整张卡，可预期

#### 热门角标
新增 `.hotBadge` 元素，绝对定位于卡片右上角：

```
+--------------------------------+
|  [icon]              [HOT]    |   ← 角标 top:0.55rem right:0.55rem
|                                |
|  Cloud Compute Resize          |
|  申请扩容/缩容...               |
|                                |
|  [tag] [tag] [tag]            |
+--------------------------------+
```

样式契约：
- 字号 `0.62rem`，字重 `700`，字距 `0.04em`，大写
- 内边距 `0 0.4rem`，最小高度 `1.05rem`，圆角 `999px`
- 边框 `1px solid rgba(86, 70, 55, 0.18)`
- 背景 `#fffaf2`，文字色 `#935d2d`
- 不可点击（`pointer-events: none`），不挡卡片本体的 hover/focus

出现条件：当 `item.tags` 数组包含字符串 `'servicePortal.tag.popular'`。

文案：新增 i18n key `servicePortal.tag.popular`（中文：「热门」，英文：「Popular」）。**注意：此 key 当前并不存在**，需要同时新增到 `messages.ts` 的 `en` 和 `zhCN` 字典。

### popular 分类过滤改造

`data/serviceCatalog.ts` 的 `listByCategory(category)` 函数：

**当前实现（依赖 hot 字段）：**
```ts
if (category === 'popular') return ITEMS.filter(i => i.hot);
```

**改造后（依赖 tag）：**
```ts
const POPULAR_TAG = 'servicePortal.tag.popular';
if (category === 'popular') return ITEMS.filter(i => i.tags.includes(POPULAR_TAG));
```

迁移规则（**已审计当前 catalog**）：
- 当前 13 条 catalog 数据中，4 条 `hot: true` 的项 **均未** 带 `'servicePortal.tag.popular'` tag
- 实现时必须：给这 4 条（`account.role-grant` / `cloud.compute-resize` / `network.vpn-access` / `support.incident`）的 `tags` 数组追加 `'servicePortal.tag.popular'`
- 删除所有 `hot: true` 字段
- `ticketTypes.ts` 中 `ServiceCatalogItem.hot?: boolean` 字段删除

### 受影响代码清单

| 文件 | 修改内容 |
|---|---|
| `locale/messages.ts` | 在 `en.servicePortal.tag` 与 `zhCN.servicePortal.tag` 段新增 `popular` 条目 |
| `ticketTypes.ts` | 删除 `ServiceCatalogItem.hot` 字段 |
| `data/serviceCatalog.ts` | 删除所有 4 处 `hot: true`；给同 4 条项目的 `tags` 数组追加 `'servicePortal.tag.popular'`；`listByCategory('popular')` 改为 tag 过滤 |
| `components/ServiceCard.tsx` | 删除 `${styles.cardWide}` 三元；新增 `.hotBadge` 渲染（条件 = `tags.includes(POPULAR_TAG)`） |
| `components/ServiceCard.module.css` | 删除 `.cardWide` 规则；新增 `.hotBadge` 规则；卡片 `position: relative` 以容纳角标 |
| `ServicePortalPage.module.css` | `.serviceGrid` 改为 `auto-fill / minmax(220px, 1fr)`；删除 `grid-auto-flow: dense` |

### i18n 改动

需要新增一个 key：
- `servicePortal.tag.popular`
  - `en`: `'Popular'`
  - `zhCN`: `'热门'`

落点：`web/src/locale/messages.ts` 的 `en` 字典里添加到 `servicePortal.tag.*` 段；TypeScript 会通过 `zhCN: typeof en` 约束强制 `zhCN` 也同步添加。

## 边界与验证

### 边界场景

| 场景 | 期望行为 |
|---|---|
| catalog 中无 popular tag 的项 | popular 分类显示「该分类暂无服务」（已有 empty state） |
| 视口宽 < 220px | 单列堆叠（auto-fill 自然降级） |
| 视口宽 ≥ 880px | 自然形成 4 列 |
| 切换分类 | C2 接入的 `<AnimatePresence mode="wait">` 已处理过渡 |
| 角标与 cardHover 状态 | 角标 `pointer-events: none`，不打断卡片 hover |

### 验证清单（视觉自检）

- [ ] 切到 popular 分类，能看到所有带 `servicePortal.tag.popular` tag 的卡（不少于 4 张）
- [ ] 切到其他分类，对应卡片正常显示
- [ ] 所有卡 `getBoundingClientRect().width` 在同一列内一致
- [ ] 角标 HOT/Popular 在中英文下均正常显示
- [ ] Tab 键聚焦每张卡的焦点框轮廓大小一致
- [ ] 鼠标 hover 任意卡，角标不消失、不变形

不引入单元测试（项目内无前端测试基础设施）。typecheck 通过 + 视觉自检通过即认为完成。

## 实施步骤

按顺序、每步独立 commit：

1. **i18n 准备**：`messages.ts` 在 `en` / `zhCN` 字典里新增 `servicePortal.tag.popular` key
2. **类型/数据迁移**：`ticketTypes.ts` 删 `hot`；`serviceCatalog.ts` 删 4 处 `hot: true`、给同 4 条加 popular tag、改 `listByCategory` 实现
3. **组件改造**：`ServiceCard.tsx` 删 `.cardWide` 三元；新增角标渲染
4. **样式改造**：`ServiceCard.module.css` 删 `.cardWide`，加 `.hotBadge`；`ServicePortalPage.module.css` 网格改 auto-fill

每步通过 `pnpm exec tsc -b` 才进入下一步。

## 不在范围

- 重新设计卡片排序逻辑（保持 catalog 数组顺序）
- 新增分类
- 修改 hot 判定逻辑之外的视觉打磨
- 引入前端测试框架

## 上游验证落空说明

本设计前期试图通过 `Read` 工具查看用户截图 `image.png` 来直接观察问题，但 `~/.claude/hooks/cbm-code-discovery-gate` 持续拦截（hook 用 PPID 做缓存键，但每次工具调用 PPID 都不同，导致 marker 永不命中）。最终基于代码定位 + 用户文字描述「点击框大小不一」推断根因 = B 节的 `.cardWide` 跨列行为，与用户口头确认。如后续 hook 修好，应用截图复核一遍。
