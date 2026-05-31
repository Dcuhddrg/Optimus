# Service Portal Header Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Optimize the `/service-portal` top panel so it becomes the only page title area, with clearer portal hierarchy and shorter copy.

**Architecture:** Keep the change local to the existing service portal frontend. Remove the duplicated `PageHeader` from `ServicePortalPage`, promote the panel title to the page `h1`, update the existing i18n description copy, and restyle only the header area in the page CSS module. No routing, service catalog, ticket, drawer, or data behavior changes.

**Tech Stack:** React 18, TypeScript, CSS Modules, Vite, npm scripts/local `node_modules` binaries, existing `useI18n` message dictionary.

**Spec:** `docs/superpowers/specs/2026-05-31-service-portal-header-design.md`

---

## Tooling Notes

This environment does not have a global `pnpm` binary. Use the checked-in `web/node_modules/` toolchain for per-task type checks and `npm run build` for the final frontend build.

Verified commands:

```bash
cd web && ./node_modules/.bin/tsc -b
cd web && npm run build
```

Do not use `npx pnpm@10` in the sandboxed environment; it may hang while trying to resolve packages over the restricted network.

## File Structure

- Modify `web/src/view/service/ServicePortalPage.tsx`
  - Remove the unused `PageHeader` import.
  - Remove the top-level `<PageHeader />` render.
  - Add a dedicated `.panelCopy` wrapper, eyebrow, and `h1` inside `.panelHeader`.
  - Keep both action controls exactly as they are: the `Link` still points to `/service-portal/tickets`; the button still calls `openGeneric`.
- Modify `web/src/view/service/ServicePortalPage.module.css`
  - Keep page, panel body, catalog area, and grid contracts intact.
  - Restyle `.portalPanel`, `.panelHeader`, new `.panelCopy`, new `.eyebrow`, `.panelHeader h1`, description, actions, and mobile layout.
- Modify `web/src/locale/messages.ts`
  - Update `en.servicePortal.description`.
  - Update `zhCN.servicePortal.description`.

No files are created. Do not edit `PageHeader`, service cards, catalog data, ticket store, drawer, router, or nav.

## Pre-Flight

- [ ] **Step 1: Confirm the worktree state**

Run:

```bash
git status --short
```

Expected:

- The only unrelated known dirty file may be `?? image.png`.
- Do not stage or modify `image.png`.
- If there are unexpected tracked modifications, inspect them before editing and do not revert user work.

- [ ] **Step 2: Record the implementation start ref**

Run:

```bash
git rev-parse --short HEAD
```

Expected:

- Save this value as `START_REF` in your notes for the final diff check.

- [ ] **Step 3: Confirm `servicePortal.description` is only used on this page**

Use codebase-memory first:

```text
codebase-memory search_code
project: home-lilin-trae-project-Optimus
pattern: servicePortal.description
path_filter: ^web/src/
mode: compact
```

Expected:

- Matches only in `web/src/view/service/ServicePortalPage.tsx`.
- If another page also uses the key, stop and revise the implementation to add a page-specific key instead of changing the shared description.

## Task 1: Update Service Portal Copy

**Files:**

- Modify: `web/src/locale/messages.ts:217-222`
- Modify: `web/src/locale/messages.ts:651-655`

- [ ] **Step 1: Replace the English description**

In `en.servicePortal.description`, replace:

```ts
description:
  'Add IT services to the current dashboard as request entry points, so business users can submit service requests and follow recent progress without leaving the workspace.',
```

with:

```ts
description:
  'Keep common IT services, request entry points, and recent progress in one place so business users can find what they need faster.',
```

- [ ] **Step 2: Replace the Chinese description**

In `zhCN.servicePortal.description`, replace:

```ts
description: '把 IT 部门可提供的服务以工单入口的形式补充到当前页面，业务用户可直接发起申请并查看最近进度。',
```

with:

```ts
description: '常用 IT 服务、申请入口和最近进度集中在这里，减少业务用户找入口的成本。',
```

- [ ] **Step 3: Run a type check**

Run:

```bash
cd web && ./node_modules/.bin/tsc -b
```

Expected:

- Command exits 0.
- No type errors from `messages.ts`.

- [ ] **Step 4: Commit the copy change**

Run:

```bash
git add web/src/locale/messages.ts
git commit -m "简化服务门户顶部说明文案" \
  -m "Constraint: Copy change is limited to servicePortal.description" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Tested: cd web && ./node_modules/.bin/tsc -b"
```

## Task 2: Make The Panel Header The Page Header

**Files:**

- Modify: `web/src/view/service/ServicePortalPage.tsx:1-92`

- [ ] **Step 1: Remove the `PageHeader` import**

Delete:

```ts
import { PageHeader } from '../../components/page/PageHeader';
```

- [ ] **Step 2: Remove the duplicated top-level `PageHeader` render**

Delete this block from the returned JSX:

```tsx
<PageHeader
  eyebrow={t('servicePortal.eyebrow')}
  title={t('servicePortal.title')}
  description={t('servicePortal.description')}
/>
```

- [ ] **Step 3: Replace the existing panel copy block**

Replace:

```tsx
<div>
  <h2>{t('servicePortal.title')}</h2>
  <p>{t('servicePortal.description')}</p>
</div>
```

with:

```tsx
<div className={styles.panelCopy}>
  <p className={styles.eyebrow}>{t('servicePortal.eyebrow')}</p>
  <h1>{t('servicePortal.title')}</h1>
  <p>{t('servicePortal.description')}</p>
</div>
```

Do not alter this actions block:

```tsx
<div className={styles.actions}>
  <Link className={styles.secondaryAction} to="/service-portal/tickets">
    {t('servicePortal.viewAll')}
  </Link>
  <button className={styles.primaryAction} type="button" onClick={openGeneric}>
    {t('servicePortal.createRequest')}
  </button>
</div>
```

- [ ] **Step 4: Run a type check**

Run:

```bash
cd web && ./node_modules/.bin/tsc -b
```

Expected:

- Command exits 0.
- No unused import error for `PageHeader`.
- No missing CSS module type errors.

- [ ] **Step 5: Commit the structure change**

Run:

```bash
git add web/src/view/service/ServicePortalPage.tsx
git commit -m "让服务门户面板承载页面标题" \
  -m "Constraint: Header structure only; actions and portal body remain unchanged" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Tested: cd web && ./node_modules/.bin/tsc -b"
```

## Task 3: Restyle The Header Panel

**Files:**

- Modify: `web/src/view/service/ServicePortalPage.module.css:11-83`

- [ ] **Step 1: Update `.portalPanel` background**

Replace the current `.portalPanel` background with a light portal gradient:

```css
.portalPanel {
  overflow: hidden;
  border: 1px solid var(--service-line);
  border-radius: 1.55rem;
  background: linear-gradient(135deg, #fffdf9 0%, #f8f2ea 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
}
```

- [ ] **Step 2: Convert `.panelHeader` to a two-column grid**

Replace the current `.panelHeader` block with:

```css
.panelHeader {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: clamp(1rem, 3vw, 2rem);
  padding: clamp(1.25rem, 2.8vw, 1.65rem);
  border-bottom: 1px solid var(--service-line);
}
```

- [ ] **Step 3: Add `.panelCopy` and `.eyebrow`**

Add after `.panelHeader`:

```css
.panelCopy {
  min-width: 0;
  max-width: 64ch;
}

.eyebrow {
  margin: 0 0 0.45rem;
  color: var(--service-accent);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
```

- [ ] **Step 4: Replace `h2` selectors with `h1` selectors**

Replace:

```css
.panelHeader h2,
.panelHeader p {
  margin: 0;
}

.panelHeader h2 {
  color: var(--service-ink);
  font-size: clamp(1.25rem, 2vw, 1.65rem);
  font-weight: 700;
  line-height: 1.12;
}

.panelHeader p {
  max-width: 62ch;
  margin-top: 0.55rem;
  color: var(--service-muted);
  line-height: 1.65;
}
```

with:

```css
.panelHeader h1,
.panelHeader p {
  margin: 0;
}

.panelHeader h1 {
  color: var(--service-ink);
  font-size: clamp(1.7rem, 2.6vw, 2.05rem);
  font-weight: 780;
  line-height: 1.08;
}

.panelCopy > p:last-child {
  margin-top: 0.72rem;
  color: var(--service-muted);
  font-size: 0.96rem;
  line-height: 1.65;
}
```

- [ ] **Step 5: Tighten actions alignment**

Update `.actions` to include stable alignment:

```css
.actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.55rem;
}
```

If the block already matches this, leave it unchanged.

- [ ] **Step 6: Add mobile stacking for the panel header**

Keep the existing `@media (max-width: 980px)` for `.panelBody`. Add a separate media query after it:

```css
@media (max-width: 720px) {
  .panelHeader {
    grid-template-columns: 1fr;
    align-items: start;
  }

  .actions {
    justify-content: flex-start;
  }

  .primaryAction,
  .secondaryAction {
    min-height: 2.5rem;
  }
}
```

- [ ] **Step 7: Run a type check**

Run:

```bash
cd web && ./node_modules/.bin/tsc -b
```

Expected:

- Command exits 0.
- CSS module class names referenced from `ServicePortalPage.tsx` are valid.

- [ ] **Step 8: Commit the style change**

Run:

```bash
git add web/src/view/service/ServicePortalPage.module.css
git commit -m "强化服务门户顶部面板层级" \
  -m "Constraint: CSS changes are limited to the service portal header area" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Tested: cd web && ./node_modules/.bin/tsc -b"
```

## Task 4: Final Build And Visual Checks

**Files:**

- Verify only; no planned edits.

- [ ] **Step 1: Run the frontend build**

Run:

```bash
cd web && npm run build
```

Expected:

- TypeScript build passes.
- Vite build passes.
- `dist/` output is generated under `web/`.

- [ ] **Step 2: Inspect the final diff**

Run:

```bash
git diff --stat START_REF..HEAD
git show --stat --oneline HEAD
```

Expected:

- Only the three intended implementation files changed across the implementation commits:
  - `web/src/locale/messages.ts`
  - `web/src/view/service/ServicePortalPage.tsx`
  - `web/src/view/service/ServicePortalPage.module.css`

- [ ] **Step 3: Perform visual self-check**

Open `/service-portal` in the existing frontend verification environment and check:

- The page has only one visible service portal title area.
- The panel header has an eyebrow, `服务门户` title, short description, and two actions.
- `查看全部服务` still navigates to `/service-portal/tickets`.
- `发起服务申请` still opens the generic request drawer.
- The catalog area, category tabs, service cards, and `MyTicketsAside` did not structurally change.
- At narrow width, header copy and actions stack without horizontal scrolling or overlap.

- [ ] **Step 4: Commit no-op verification if needed**

No commit is required if verification does not change files. If build output or local artifacts appear, do not stage them.
