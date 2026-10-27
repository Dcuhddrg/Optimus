# Service Portal · 顶部面板展示优化设计

> 日期：2026-05-31
> 状态：Draft，待 spec 复审
> 上游依赖：[`2026-05-28-service-portal-design.md`](2026-05-28-service-portal-design.md)
> 页面范围：`/service-portal` 首页顶部面板

## 背景

服务门户首页当前同时渲染通用 `PageHeader` 和 `portalPanel` 内部标题区，两处都展示「服务门户」与同一段说明文案。截图中可见顶部面板留白偏大、标题说明重复、按钮与文案之间的视觉层级不够明确。

本次优化只处理截图中的顶部面板展示，让它成为页面的主标题入口。服务目录卡片、分类 tab、我的工单侧栏、申请 Drawer、工单列表和数据逻辑都保持不变。

## 目标

- 消除 `/service-portal` 首屏标题和说明的重复展示。
- 保留顶部面板作为「服务门户」的大入口视觉。
- 强化标题、说明、按钮的层级关系。
- 使用更短的门户型说明文案，降低首屏阅读负担。
- 保持现有交互行为：`查看全部服务` 仍跳转全部工单页面，`发起服务申请` 仍打开通用申请 Drawer。

## 非目标

- 不改服务卡片尺寸、排序、标签、热门角标或分类过滤逻辑。
- 不改 `MyTicketsAside`、`RequestDrawer`、`RequestForm`、工单列表或详情页。
- 不新增导航、路由、后端接口或依赖。
- 不新增统计 chip 或动态摘要数据。
- 不调整全局 `PageHeader` 组件，避免影响其他页面。

## 已选方案

采用「强化门户感」方案：

- 删除 `ServicePortalPage` 中顶部的通用 `PageHeader` 调用。
- 让 `portalPanel` 的 `.panelHeader` 成为服务门户首页的主标题区。
- 在 `.panelHeader` 文案区新增轻量 eyebrow：`Service portal` / `服务门户` 对应现有 `servicePortal.eyebrow`。
- 标题继续使用 `servicePortal.title`。
- 说明文案改为更短的门户型文案。
- 面板头部使用轻量渐变背景和更明确的按钮布局，保留现有色系。

## 文案

更新 `servicePortal.description`：

| Locale | 文案 |
| --- | --- |
| `zhCN` | `常用 IT 服务、申请入口和最近进度集中在这里，减少业务用户找入口的成本。` |
| `en` | `Keep common IT services, request entry points, and recent progress in one place so business users can find what they need faster.` |

说明：

- 中文文案沿用用户确认的短文案。
- 英文同步表达相同含义，不新增未被确认的业务承诺。
- 因为 `servicePortal.description` 也可能被其他服务门户子页面引用，实施前需要确认实际引用点；本设计预期当前只影响服务门户首页入口文案。如发现该 key 被工单列表、详情页等页面复用，应改为新增 `servicePortal.heroDescription`，避免扩大影响面。

## 组件结构

### `ServicePortalPage.tsx`

当前结构：

```tsx
<div className={styles.page}>
  <PageHeader
    eyebrow={t('servicePortal.eyebrow')}
    title={t('servicePortal.title')}
    description={t('servicePortal.description')}
  />

  <section className={styles.portalPanel}>
    <div className={styles.panelHeader}>
      ...
    </div>
    ...
  </section>
</div>
```

改造后：

```tsx
<div className={styles.page}>
  <section className={styles.portalPanel}>
    <div className={styles.panelHeader}>
      <div className={styles.panelCopy}>
        <p className={styles.eyebrow}>{t('servicePortal.eyebrow')}</p>
        <h1>{t('servicePortal.title')}</h1>
        <p>{t('servicePortal.description')}</p>
      </div>
      <div className={styles.actions}>...</div>
    </div>
    ...
  </section>
</div>
```

设计要求：

- 使用 `h1` 保持页面主标题语义；不再在同一页面重复另一个 `h1`。
- `PageHeader` import 删除，只影响当前页面。
- 按钮行为和文案 key 不变。

## 样式合同

### 面板

`.portalPanel` 继续作为整个服务门户首页内容容器：

- 保持现有边框与圆角体系。
- 背景保留暖白基调，可用轻量 `linear-gradient(135deg, #fffdf9, #f8f2ea)` 强化入口感。
- 不调整 `.panelBody` 和 `.serviceGrid` 的布局合同。

### 面板头部

`.panelHeader`：

- 桌面端使用两列布局：左侧标题文案 `minmax(0, 1fr)`，右侧按钮 `auto`。
- 对齐方式为垂直居中，避免按钮悬浮在大块留白中。
- 内边距略大于当前普通分隔头部，但不超过截图现有视觉高度。
- 底部边框保留，区分头部和下方目录区域。

`.panelCopy`：

- 限制最大宽度，避免长行影响可读性。
- `eyebrow` 小字号、加粗、使用服务门户主色。
- `h1` 比当前 `h2` 更明确，但保持平台风格，不使用过大的 hero 字号。
- 说明文案行高保持舒适，最大宽度约 `64ch`。

`.actions`：

- 保持两个按钮横向排列。
- 主按钮继续使用 `--service-accent`。
- 次按钮继续使用浅底 + 边框。
- 桌面端右对齐，小屏端左对齐并允许换行。

### 响应式

- `max-width: 720px` 左右时，`.panelHeader` 改为单列。
- 按钮容器宽度可铺满，按钮可换行。
- 不出现文字溢出、按钮遮挡或横向滚动。

## 受影响文件

| 文件 | 预期修改 |
| --- | --- |
| `web/src/view/service/ServicePortalPage.tsx` | 删除顶部 `PageHeader` 调用和 import；在 `.panelHeader` 内加入 eyebrow 与 `h1` 结构 |
| `web/src/view/service/ServicePortalPage.module.css` | 调整 `.portalPanel`、`.panelHeader`、标题、说明、按钮和小屏样式 |
| `web/src/locale/messages.ts` | 更新 `servicePortal.description` 的中英文文案，或在引用面过大时新增首页专用 key |

## 验证

实现后执行：

- `cd web && pnpm build`

视觉自检：

- `/service-portal` 顶部只出现一次主标题和说明。
- 顶部面板符合方案 B：有 eyebrow、标题、短说明、右侧两个按钮。
- 点击 `查看全部服务` 路径不变。
- 点击 `发起服务申请` 仍打开通用申请 Drawer。
- 桌面宽度下按钮位于右侧且不挤压文案。
- 窄屏下标题、说明、按钮堆叠正常，无横向滚动。
- 服务卡片区、分类 tab、我的工单侧栏没有结构变化。

## 风险与约束

- `servicePortal.description` 若被其他页面共享，直接改文案可能影响非首页位置。实施前必须确认引用面，必要时改用新增首页专用 key。
- 删除 `PageHeader` 会改变 `/service-portal` 与其他页面的头部一致性；这是本次设计的有意取舍，因为该页面已经有专属门户面板。
- 本次只做展示优化，不补充真实服务数量或最近进度统计，避免把 mock 数据包装成业务指标。
