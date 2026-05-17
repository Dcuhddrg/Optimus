# SP4 资产页补全 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans. Steps use checkbox (`- [ ]`) syntax.

**Goal:** 把 CloudAssetsPage 从最小骨架补齐到 mockup 完整体验：补列、补传统过滤栏、加分页、软删除淡化、分层显示账号继承 tag 与云原生 tag。

**Architecture:** 纯前端补丁，**后端不动**。当前 `/api/v1/cloud/assets` 接口已经支持 `accountId / provider / region / resourceType / status / keyword / page / pageSize / tag` 全部 query 参数；前端 API client 和页面没把这些参数 wire 上。本 plan 把它们 wire 通，并把每个资产的 `tagsJson`（云原生 tag）单独显示一列。

**Tech Stack:** React 18 + TypeScript + Vite + CSS Modules，无新依赖。

**Spec reference:**
- 设计文档：`docs/superpowers/specs/2026-05-17-cloud-account-tags-design.md` §"与 cloud_assets.tags_json 区分"
- mockup：`environment-tags-mockup.html` 屏幕 3
- 当前实现：`web/src/view/cloud/CloudAssetsPage.tsx`（88 行，仅骨架）

**对照 mockup 的 4 类缺口：**

| 缺口 | 解决方式 |
|---|---|
| 表格缺 region / status / lastSeen 列；软删除资产无淡化 | Task 1 |
| 没有传统过滤栏（账号/厂商/region/类型/关键字） | Task 2 |
| 没有分页 UI | Task 3 |
| 云原生 tag (`cloud_assets.tags_json`) 没有任何出口 | Task 4 |

---

## File Structure

**修改：**
- `web/src/api/cloud.ts` — `listCloudAssets` 接收完整过滤参数（不仅是 tagFilters）
- `web/src/view/cloud/CloudAssetsPage.tsx` — 加列、加过滤栏、加分页、加云原生 tag 列
- `web/src/view/cloud/CloudAssetsPage.module.css` — 调表格列宽 + 软删除样式
- `web/src/locale/messages.ts` — 新增 i18n keys

**新增：** 无（不引入新组件，过滤栏直接 inline 写在 page 里；如果将来要复用再抽组件）

---

### Task 1: 资产表格列补全 + 软删除淡化

**Files:**
- Modify: `web/src/view/cloud/CloudAssetsPage.tsx`
- Modify: `web/src/view/cloud/CloudAssetsPage.module.css`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 1.1: 加 i18n keys**

`web/src/locale/messages.ts`，找到 `cloud.assets` 段（EN 和 ZH 各一处），追加 keys（既保留现有 name/resourceId/type/account/inheritedTags，再加新的）：

EN `cloud.assets`：
```ts
region: 'Region',
status: 'Status',
lastSeen: 'Last seen',
nativeTags: 'Native tags',
deletedBadge: 'Deleted',
deletedTitle: 'This asset was last seen before the most recent sync and may be deleted upstream.',
```

ZH `cloud.assets`：
```ts
region: '区域',
status: '状态',
lastSeen: '最近发现',
nativeTags: '云原生 tag',
deletedBadge: '已下线',
deletedTitle: '该资产已从云上消失，软删除保留供历史查询。',
```

- [ ] **Step 1.2: 在 CloudAssetsPage 表格里加 4 列**

把现有 `tableHeader` 改成 7 列（保留前 4 列，加 region/status/lastSeen 列，把 inheritedTags 改名为更短的"账号 tag" 用 `t('cloud.assets.inheritedTags')` 不变；再加 `nativeTags` 一列由 Task 4 接管，本 task 先占位）：

```tsx
<div className={styles.tableHeader} aria-hidden="true">
  <span>{t('cloud.assets.name')}</span>
  <span>{t('cloud.assets.resourceId')}</span>
  <span>{t('cloud.assets.type')}</span>
  <span>{t('cloud.assets.region')}</span>
  <span>{t('cloud.assets.status')}</span>
  <span>{t('cloud.assets.account')}</span>
  <span>{t('cloud.assets.lastSeen')}</span>
  <span>{t('cloud.assets.inheritedTags')}</span>
</div>
```

行渲染（替换现有 `.row` 内的 5 个子元素）：

```tsx
{assets.map((asset) => {
  const account = accountsById[asset.accountId];
  const accountTags = account?.tags ?? {};
  const isDeleted = asset.deletedAt !== null;
  return (
    <article
      className={[styles.row, isDeleted ? styles.rowDeleted : ''].filter(Boolean).join(' ')}
      key={asset.id}
      title={isDeleted ? t('cloud.assets.deletedTitle') : undefined}
    >
      <strong>
        {asset.name}
        {isDeleted ? <span className={styles.deletedBadge}>{t('cloud.assets.deletedBadge')}</span> : null}
      </strong>
      <code>{asset.resourceId}</code>
      <span>{asset.resourceType}</span>
      <span>{asset.region}</span>
      <span className={styles.status}>{asset.status || '—'}</span>
      <span>{account?.name ?? `#${asset.accountId}`}</span>
      <span className={styles.timestamp}>{formatTimestamp(asset.lastSeenAt)}</span>
      <div className={styles.tags}>
        {Object.entries(accountTags).length === 0 ? (
          <span className={styles.untagged}>{t('cloud.accounts.untagged')}</span>
        ) : (
          Object.entries(accountTags).map(([key, value]) => (
            <TagChip key={key} tagKey={key} value={value} />
          ))
        )}
      </div>
    </article>
  );
})}
```

在文件底部追加辅助函数：

```tsx
function formatTimestamp(iso: string | null): string {
  if (!iso) return '—';
  try {
    const date = new Date(iso);
    return date.toLocaleString(undefined, {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return iso;
  }
}
```

> Note: `CloudAsset` 类型里 `lastSeenAt` 已经是 `string`（不是 `string | null`）—— 看一眼 `types/cloud.ts`。如果是非 null 的 `string`，把 `iso: string | null` 改成 `iso: string`，并在 isDeleted 判断保持原状（`deletedAt: string | null` 才是 nullable）。**实施者检查 types/cloud.ts 后按实际类型对齐**。

- [ ] **Step 1.3: 调 CSS 适配 8 列 + 软删除样式**

`CloudAssetsPage.module.css`，替换 `.tableHeader, .row` 部分：

```css
.tableHeader,
.row {
  display: grid;
  grid-template-columns:
    minmax(8rem, 1.2fr)   /* name */
    minmax(9rem, 1.2fr)   /* resourceId */
    minmax(6rem, 0.8fr)   /* type */
    minmax(6rem, 0.8fr)   /* region */
    minmax(5rem, 0.6fr)   /* status */
    minmax(7rem, 1fr)     /* account */
    minmax(8rem, 0.9fr)   /* lastSeen */
    minmax(10rem, 1.5fr); /* tags */
  gap: 10px;
  align-items: center;
  padding: 10px 12px;
}
```

新增样式：

```css
.rowDeleted {
  opacity: 0.55;
  background: rgba(255, 255, 255, 0.42);
}

.rowDeleted strong {
  text-decoration: line-through;
  text-decoration-thickness: 1px;
}

.deletedBadge {
  display: inline-block;
  margin-left: 6px;
  padding: 1px 7px;
  border-radius: 999px;
  background: rgba(110, 93, 81, 0.16);
  color: var(--shell-muted);
  font-size: 0.66rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  vertical-align: middle;
}

.status {
  color: var(--shell-muted);
  font-size: 0.84rem;
  font-weight: 600;
}

.timestamp {
  color: var(--shell-muted);
  font-size: 0.82rem;
  font-variant-numeric: tabular-nums;
}
```

- [ ] **Step 1.4: 构建 + 提交**

```bash
cd /home/lilin/trae-project/Optimus/web && npm run build
```

Expected: clean.

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/view/cloud/CloudAssetsPage.tsx web/src/view/cloud/CloudAssetsPage.module.css web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
SP4 资产页补丁 1：补 region/status/lastSeen 列与软删除淡化

Constraint: 不动后端；软删除资产用透明度+删除线+badge 三重信号
Confidence: high
Scope-risk: narrow
Tested: npm run build
EOF
)"
```

---

### Task 2: 资产页传统过滤栏（账号 / 厂商 / region / 类型 / 关键字）

**Files:**
- Modify: `web/src/api/cloud.ts`
- Modify: `web/src/types/cloud.ts`
- Modify: `web/src/view/cloud/CloudAssetsPage.tsx`
- Modify: `web/src/view/cloud/CloudAssetsPage.module.css`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 2.1: 扩展 types/cloud.ts**

新增 `CloudAssetFilters` 类型（在 `CloudAssetPage` 之后）：

```ts
export interface CloudAssetFilters {
  accountId?: number;
  provider?: string;
  region?: string;
  resourceType?: string;
  status?: string;
  keyword?: string;
  page?: number;
  pageSize?: number;
  tag?: string[];
}
```

- [ ] **Step 2.2: 改 api/cloud.ts 的 listCloudAssets 接收完整过滤**

替换现有 `listCloudAssets`：

```ts
import type {
  CloudAccount,
  CloudAssetFilters,
  CloudAssetPage,
  CreateCloudAccountPayload,
  UpdateCloudAccountTagsPayload,
} from '../types/cloud';

function buildAssetQuery(filters: CloudAssetFilters): string {
  const params = new URLSearchParams();
  if (filters.accountId !== undefined) params.set('accountId', String(filters.accountId));
  if (filters.provider) params.set('provider', filters.provider);
  if (filters.region) params.set('region', filters.region);
  if (filters.resourceType) params.set('resourceType', filters.resourceType);
  if (filters.status) params.set('status', filters.status);
  if (filters.keyword) params.set('keyword', filters.keyword);
  if (filters.page !== undefined) params.set('page', String(filters.page));
  if (filters.pageSize !== undefined) params.set('pageSize', String(filters.pageSize));
  if (filters.tag) {
    for (const t of filters.tag) {
      params.append('tag', t);
    }
  }
  const qs = params.toString();
  return qs ? `?${qs}` : '';
}

export function listCloudAssets(filters: CloudAssetFilters = {}) {
  return apiRequest<CloudAssetPage>(`/cloud/assets${buildAssetQuery(filters)}`);
}
```

> 注意：旧签名 `listCloudAssets(tagFilters: string[])` 被替换。**搜全仓库 `listCloudAssets(`** 调用方，确保都传新格式。当前应只有 `CloudAssetsPage.tsx` 一个调用方，本 task 会更新它。

- [ ] **Step 2.3: 在 CloudAssetsPage 里加过滤栏 UI + state**

在 page 顶部新增 state 与计算属性：

```tsx
import { ChangeEvent, useEffect, useMemo, useState } from 'react';
// ...

interface BasicFilters {
  accountId: string; // '' = all
  provider: string;
  region: string;
  resourceType: string;
  status: string;
  keyword: string;
}

const emptyBasicFilters: BasicFilters = {
  accountId: '',
  provider: '',
  region: '',
  resourceType: '',
  status: '',
  keyword: '',
};

export function CloudAssetsPage() {
  const { t } = useI18n();
  const [assets, setAssets] = useState<CloudAsset[]>([]);
  const [total, setTotal] = useState(0);
  const [accountsById, setAccountsById] = useState<Record<number, CloudAccount>>({});
  const [accountsList, setAccountsList] = useState<CloudAccount[]>([]);
  const [tagFilters, setTagFilters] = useState<string[]>([]);
  const [basic, setBasic] = useState<BasicFilters>(emptyBasicFilters);
  const [error, setError] = useState<string | null>(null);

  const providers = useMemo(
    () => Array.from(new Set(accountsList.map((a) => a.provider))).sort(),
    [accountsList],
  );
  const resourceTypes = useMemo(
    () => Array.from(new Set(assets.map((a) => a.resourceType))).filter(Boolean).sort(),
    [assets],
  );

  async function load() {
    try {
      const [assetsResult, accountsResult] = await Promise.all([
        listCloudAssets({
          accountId: basic.accountId ? Number(basic.accountId) : undefined,
          provider: basic.provider || undefined,
          region: basic.region || undefined,
          resourceType: basic.resourceType || undefined,
          status: basic.status || undefined,
          keyword: basic.keyword || undefined,
          tag: tagFilters.length > 0 ? tagFilters : undefined,
        }),
        listCloudAccounts(),
      ]);
      setAssets(assetsResult.items);
      setTotal(assetsResult.total);
      setAccountsList(accountsResult.items);
      const indexed: Record<number, CloudAccount> = {};
      for (const account of accountsResult.items) indexed[account.id] = account;
      setAccountsById(indexed);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.assets.loadFailed'));
    }
  }

  useEffect(() => { void load(); }, [tagFilters, basic]);

  function updateBasic(key: keyof BasicFilters, event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) {
    const value = event.target.value;
    setBasic((prev) => ({ ...prev, [key]: value }));
  }

  // ... existing JSX, with filter row inserted between TagFilter and table
}
```

在 JSX 里 `<TagFilter ...>` 后插入过滤栏：

```tsx
<div className={styles.filterRow}>
  <select value={basic.accountId} onChange={(e) => updateBasic('accountId', e)}>
    <option value="">{t('cloud.assets.allAccounts')}</option>
    {accountsList.map((account) => (
      <option key={account.id} value={String(account.id)}>{account.name}</option>
    ))}
  </select>
  <select value={basic.provider} onChange={(e) => updateBasic('provider', e)}>
    <option value="">{t('cloud.assets.allProviders')}</option>
    {providers.map((p) => <option key={p} value={p}>{p}</option>)}
  </select>
  <input
    value={basic.region}
    onChange={(e) => updateBasic('region', e)}
    placeholder={t('cloud.assets.regionPlaceholder')}
  />
  <select value={basic.resourceType} onChange={(e) => updateBasic('resourceType', e)}>
    <option value="">{t('cloud.assets.allTypes')}</option>
    {resourceTypes.map((tp) => <option key={tp} value={tp}>{tp}</option>)}
  </select>
  <input
    value={basic.status}
    onChange={(e) => updateBasic('status', e)}
    placeholder={t('cloud.assets.statusPlaceholder')}
  />
  <input
    value={basic.keyword}
    onChange={(e) => updateBasic('keyword', e)}
    placeholder={t('cloud.assets.keywordPlaceholder')}
  />
</div>
```

- [ ] **Step 2.4: i18n keys**

EN `cloud.assets` 追加：
```ts
allAccounts: 'All accounts',
allProviders: 'All providers',
allTypes: 'All types',
regionPlaceholder: 'Region',
statusPlaceholder: 'Status',
keywordPlaceholder: 'Search by name',
```

ZH 追加：
```ts
allAccounts: '所有账号',
allProviders: '所有厂商',
allTypes: '所有类型',
regionPlaceholder: '区域',
statusPlaceholder: '状态',
keywordPlaceholder: '搜索名称',
```

- [ ] **Step 2.5: CSS 加 filterRow 样式**

`CloudAssetsPage.module.css` 追加：

```css
.filterRow {
  display: grid;
  grid-template-columns: repeat(6, minmax(8rem, 1fr));
  gap: 8px;
  padding: 10px 12px;
  border: 1px solid var(--shell-border);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.5);
}

.filterRow select,
.filterRow input {
  min-height: 34px;
  padding: 0 10px;
  border: 1px solid var(--shell-border);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.78);
  color: var(--text-primary);
  font-size: 0.84rem;
}

.filterRow select:focus,
.filterRow input:focus {
  border-color: rgba(47, 95, 89, 0.34);
  outline: 2px solid rgba(47, 95, 89, 0.12);
  outline-offset: 1px;
}

@media (max-width: 900px) {
  .filterRow {
    grid-template-columns: repeat(2, 1fr);
  }
}
```

- [ ] **Step 2.6: 构建 + 提交**

```bash
cd /home/lilin/trae-project/Optimus/web && npm run build
```

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/api/cloud.ts web/src/types/cloud.ts web/src/view/cloud/CloudAssetsPage.tsx web/src/view/cloud/CloudAssetsPage.module.css web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
SP4 资产页补丁 2：传统过滤栏（账号/厂商/region/类型/状态/关键字）

Constraint: listCloudAssets 改为接收 CloudAssetFilters 对象；厂商和类型下拉从已有数据动态聚合
Confidence: high
Scope-risk: narrow（API 签名变化但仅一处调用方）
Tested: npm run build
EOF
)"
```

---

### Task 3: 分页

**Files:**
- Modify: `web/src/view/cloud/CloudAssetsPage.tsx`
- Modify: `web/src/view/cloud/CloudAssetsPage.module.css`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 3.1: 加分页 state + 计算**

在 page 顶部 state 加：

```tsx
const [page, setPage] = useState(1);
const [pageSize, setPageSize] = useState(20);
const totalPages = Math.max(1, Math.ceil(total / pageSize));
```

修改 `load()` 把 page/pageSize 塞入 filter：

```tsx
listCloudAssets({
  accountId: basic.accountId ? Number(basic.accountId) : undefined,
  provider: basic.provider || undefined,
  region: basic.region || undefined,
  resourceType: basic.resourceType || undefined,
  status: basic.status || undefined,
  keyword: basic.keyword || undefined,
  tag: tagFilters.length > 0 ? tagFilters : undefined,
  page,
  pageSize,
}),
```

修改 useEffect 依赖：

```tsx
useEffect(() => { void load(); }, [tagFilters, basic, page, pageSize]);
```

**当过滤变化时把 page 重置为 1**：在 `setBasic` 和 `setTagFilters` 包装一层：

```tsx
function updateBasic(key: keyof BasicFilters, event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) {
  const value = event.target.value;
  setBasic((prev) => ({ ...prev, [key]: value }));
  setPage(1);
}

function handleTagFiltersChange(next: string[]) {
  setTagFilters(next);
  setPage(1);
}
```

把 `<TagFilter onChange={setTagFilters} ...>` 改成 `<TagFilter onChange={handleTagFiltersChange} ...>`。

- [ ] **Step 3.2: 在表格底部加分页控件**

在 `</div>` 关闭表格 list 后追加：

```tsx
<div className={styles.pagination}>
  <span className={styles.paginationSummary}>
    {t('cloud.assets.totalCount', String(total))}
  </span>
  <div className={styles.paginationControls}>
    <button
      type="button"
      disabled={page <= 1}
      onClick={() => setPage((p) => Math.max(1, p - 1))}
    >
      {t('cloud.assets.prevPage')}
    </button>
    <span className={styles.paginationPosition}>
      {page} / {totalPages}
    </span>
    <button
      type="button"
      disabled={page >= totalPages}
      onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
    >
      {t('cloud.assets.nextPage')}
    </button>
    <select
      value={String(pageSize)}
      onChange={(e) => { setPageSize(Number(e.target.value)); setPage(1); }}
    >
      <option value="20">20 / {t('cloud.assets.perPage')}</option>
      <option value="50">50 / {t('cloud.assets.perPage')}</option>
      <option value="100">100 / {t('cloud.assets.perPage')}</option>
    </select>
  </div>
</div>
```

> Note: `useI18n` 的 `t()` 是否支持参数化（`t('cloud.assets.totalCount', String(total))`）取决于实现。**实施者打开 `web/src/locale/useI18n.ts` 检查**：
> - 若支持参数：直接用
> - 若不支持：把 `totalCount` 拆成两段 `total: '共'` + `entries: '条'`，组合显示成 `{t('total')} {total} {t('entries')}`

- [ ] **Step 3.3: i18n keys**

EN `cloud.assets`：
```ts
totalCount: 'Total {0}',
prevPage: '‹',
nextPage: '›',
perPage: 'per page',
```

ZH：
```ts
totalCount: '共 {0} 条',
prevPage: '‹',
nextPage: '›',
perPage: '每页',
```

如果 `t()` 不支持 `{0}` 占位，按 Step 3.2 备注用拆分 keys。

- [ ] **Step 3.4: CSS 分页样式**

`CloudAssetsPage.module.css` 追加：

```css
.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 4px 0;
  color: var(--shell-muted);
  font-size: 0.84rem;
}

.paginationControls {
  display: flex;
  align-items: center;
  gap: 6px;
}

.paginationControls button {
  min-height: 32px;
  padding: 0 12px;
  border: 1px solid var(--shell-border);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.72);
  color: var(--shell-control-active-bg);
  font-weight: 700;
  cursor: pointer;
}

.paginationControls button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.paginationControls select {
  min-height: 32px;
  margin-left: 8px;
  padding: 0 8px;
  border: 1px solid var(--shell-border);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.72);
}

.paginationPosition {
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
  font-weight: 700;
}
```

- [ ] **Step 3.5: 构建 + 提交**

```bash
cd /home/lilin/trae-project/Optimus/web && npm run build
```

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/view/cloud/CloudAssetsPage.tsx web/src/view/cloud/CloudAssetsPage.module.css web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
SP4 资产页补丁 3：分页控件

Constraint: 过滤/tag 变化时页码自动重置到 1；每页 20/50/100 可选
Confidence: high
Scope-risk: narrow
Tested: npm run build
EOF
)"
```

---

### Task 4: 云原生 tag 列（与账号继承 tag 分层显示）

**Files:**
- Modify: `web/src/view/cloud/CloudAssetsPage.tsx`
- Modify: `web/src/view/cloud/CloudAssetsPage.module.css`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 4.1: 加云原生 tag 列**

`CloudAsset.tagsJson` 是 syncer 拉来的云厂商原生 tag（JSON 字符串）。需要在前端 parse 并展示。

在 `CloudAssetsPage.tsx` 文件底部加辅助函数：

```tsx
function parseNativeTags(tagsJson: string): Record<string, string> {
  if (!tagsJson) return {};
  try {
    const parsed = JSON.parse(tagsJson);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      const out: Record<string, string> = {};
      for (const [k, v] of Object.entries(parsed)) {
        if (typeof v === 'string') out[k] = v;
        else if (v != null) out[k] = String(v);
      }
      return out;
    }
  } catch {
    // ignore parse errors; tags_json was not a JSON object
  }
  return {};
}
```

在 `tableHeader` 末尾新增一列：

```tsx
<span>{t('cloud.assets.inheritedTags')}</span>
<span>{t('cloud.assets.nativeTags')}</span>
```

并把表格 `.tableHeader, .row` 改成 9 列（grid-template-columns 加一列：见 Step 4.2）。

在 row 渲染里追加：

```tsx
const nativeTags = parseNativeTags(asset.tagsJson);
// ... existing JSX
<div className={styles.tags}>
  {Object.entries(nativeTags).length === 0 ? (
    <span className={styles.untagged}>{t('cloud.assets.noNativeTags')}</span>
  ) : (
    Object.entries(nativeTags).map(([key, value]) => (
      <TagChip key={key} tagKey={key} value={value} />
    ))
  )}
</div>
```

> 注意：`TagChip` 对 `environment` key 会上色——这是用户 tag 的约定，**云原生 tag 里如果也有 environment key 会复用同样的着色**。这是预期：云原生 tag 用 environment=prod 也应该有同样的视觉提示。如果你想视觉区分（云原生 tag 用更柔和的色），未来再加 `variant="native"` 之类 prop。本 task 先不区分。

- [ ] **Step 4.2: CSS 改为 9 列**

替换 Task 1 落地的 grid-template-columns：

```css
.tableHeader,
.row {
  display: grid;
  grid-template-columns:
    minmax(7rem, 1.1fr)   /* name */
    minmax(8rem, 1.1fr)   /* resourceId */
    minmax(5rem, 0.7fr)   /* type */
    minmax(5rem, 0.7fr)   /* region */
    minmax(4.5rem, 0.5fr) /* status */
    minmax(6rem, 0.9fr)   /* account */
    minmax(7rem, 0.85fr)  /* lastSeen */
    minmax(9rem, 1.3fr)   /* inheritedTags */
    minmax(9rem, 1.3fr);  /* nativeTags */
  gap: 8px;
  align-items: center;
  padding: 10px 12px;
}
```

- [ ] **Step 4.3: i18n keys**

EN `cloud.assets`：
```ts
noNativeTags: 'No native tags',
```

ZH：
```ts
noNativeTags: '无云原生 tag',
```

`nativeTags` 标题列已经在 Task 1 加了，无需重复。

- [ ] **Step 4.4: 构建 + 提交**

```bash
cd /home/lilin/trae-project/Optimus/web && npm run build
```

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/view/cloud/CloudAssetsPage.tsx web/src/view/cloud/CloudAssetsPage.module.css web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
SP4 资产页补丁 4：云原生 tag 分层显示

Constraint: parseNativeTags 防御性 try/catch；非对象/非 string 值跳过
Confidence: high
Scope-risk: narrow
Tested: npm run build
EOF
)"
```

---

## Self-Review Checklist

**Spec coverage：**
- ✅ Region / Status / LastSeen 列（Task 1）
- ✅ 软删除淡化 + badge（Task 1）
- ✅ 账号 / 厂商 / region / 类型 / 状态 / 关键字 过滤（Task 2）
- ✅ 分页控件（Task 3）
- ✅ 云原生 tag 分层显示（Task 4）

**Placeholder scan：**
- Step 1.2 末尾 "实施者检查 types/cloud.ts 后按实际类型对齐" — 这不是 placeholder，是显式 invariant：实际类型由现有代码决定，实施者读 1 行确认即可
- Step 3.2 末尾 "若不支持参数化" 是 conditional fallback，给了具体替代实现，不是 placeholder

**Type consistency：**
- `CloudAssetFilters` 在 Task 2 定义，Task 3 复用并新增 page/pageSize
- `basic: BasicFilters` 在 Task 2 内部 state；Task 3 加 `page/pageSize` 时 page 重置逻辑覆盖 basic 的所有 setter
- `parseNativeTags` 在 Task 4 加，仅在 Task 4 调用
- TagChip 同一组件同时用于 inherited tags 和 native tags（共用配色规则）

**外部依赖：**
- 后端 `/api/v1/cloud/assets` 已支持全部 query 参数（无需后端改动）
- `apiRequest` 已支持 URLSearchParams 拼接（Task 2 用 URLSearchParams 而不是手工字符串拼接）

---

## Out of Scope

- 资产 detail 抽屉 / 详情页
- 排序（按 lastSeen / name 排序）
- 列宽用户自定义
- 导出 CSV
- 资产级用户 tag（仍仅账号级；spec 已声明）
- 区分 "云原生 tag" 与 "用户 tag" 的视觉风格变体（统一用 TagChip，environment 共享配色）
- 移除/恢复软删除资产的 UI 操作（保持 deleted_at 只读展示）

---

## Run-Through Walkthrough

实施完 4 个 task 后，端到端调用链：

1. 用户切到 prod 项目，访问 `/cloud/assets`
2. 页面加载：`Promise.all([listCloudAssets({page:1, pageSize:20}), listCloudAccounts()])`
3. 后端按 ProjectScope 中间件过滤可访问项目，返回 page+total
4. 表格渲染 9 列：name / resourceId / type / region / status / account / lastSeen / inherited / native
5. 软删除资产（`deletedAt !== null`）行透明度降为 0.55、名称加删除线、加 "已下线" badge
6. 用户在过滤栏选账号 → setBasic + setPage(1) → useEffect → 重拉
7. 用户在 TagFilter 点 environment=prod 芯片 → setTagFilters + setPage(1) → useEffect → 重拉
8. 用户点底部 `›` → setPage(p+1) → useEffect → 重拉新页
9. 用户改每页 50 → setPageSize(50) + setPage(1) → useEffect → 重拉

---

## 备注

本 plan 与 SP4 主 plan 互相独立，但前置依赖 SP4 已完成（CloudAssetsPage 基础骨架在 `e76d8ab` 已落地）。可以单独串行实施这 4 个 task，每个独立 commit。
