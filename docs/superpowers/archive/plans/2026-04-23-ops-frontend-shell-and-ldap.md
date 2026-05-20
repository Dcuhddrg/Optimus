# Ops Frontend Shell And LDAP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first React + TypeScript frontend shell under `web/` with only `Dashboard` and `System > LDAP`, where `Dashboard` is an empty-state page and `LDAP` is a static configuration-page shell.

**Architecture:** The first frontend iteration stays intentionally narrow. The app shell provides only two top-level navigation entries, uses a gin-vue-admin-inspired directory organization, and renders static productized pages without real LDAP logic or a fake data-heavy dashboard. The implementation separates app bootstrapping, shell layout, page modules, and shared UI primitives so later expansion does not require rewiring the first version.

**Tech Stack:** React, TypeScript, Vite, React Router, CSS Modules or scoped stylesheet files, static mock content only

---

## File Map

### Create

- `web/package.json` — frontend package manifest and scripts
- `web/tsconfig.json` — TypeScript config
- `web/tsconfig.node.json` — Vite/Node TS config
- `web/vite.config.ts` — Vite app config
- `web/index.html` — app entry HTML
- `web/src/main.tsx` — React bootstrap
- `web/src/App.tsx` — root router outlet
- `web/src/style/index.css` — global theme tokens and reset
- `web/src/router/index.tsx` — router definition
- `web/src/layout/AppShell.tsx` — two-entry app shell
- `web/src/layout/AppShell.module.css` — shell styling
- `web/src/view/dashboard/DashboardPage.tsx` — dashboard empty state page
- `web/src/view/dashboard/DashboardPage.module.css` — dashboard page styling
- `web/src/view/system/SystemLdapPage.tsx` — LDAP static shell page
- `web/src/view/system/SystemLdapPage.module.css` — LDAP page styling
- `web/src/components/nav/SideNav.tsx` — shell navigation component
- `web/src/components/nav/SideNav.module.css` — nav styling
- `web/src/components/page/PageHeader.tsx` — shared page header component
- `web/src/components/page/PageHeader.module.css` — shared header styling
- `web/src/components/state/EmptyState.tsx` — shared empty state component
- `web/src/components/state/EmptyState.module.css` — empty state styling

### Modify

- `README.md` — optionally add one short frontend startup note only after the app is runnable

### Do Not Modify

- `server/**` — no backend changes
- `.codex/**` — no Codex rule changes in this feature
- `docs/architecture/**` — no architecture doc changes in this feature

## Task 1: Initialize the React + TypeScript frontend workspace

**Files:**
- Create: `web/package.json`
- Create: `web/tsconfig.json`
- Create: `web/tsconfig.node.json`
- Create: `web/vite.config.ts`
- Create: `web/index.html`

- [ ] **Step 1: Write the failing frontend manifest check**

Create a temporary checklist in your notes:

```text
- web/package.json exists
- web/index.html exists
- web/vite.config.ts exists
- TypeScript config files exist
```

- [ ] **Step 2: Verify the workspace currently fails the checklist**

Run: `find web -maxdepth 1 -type f | sort`
Expected: FAIL the checklist because the frontend app bootstrap files do not exist yet.

- [ ] **Step 3: Write minimal workspace files**

Create:

- `package.json` with `dev`, `build`, and `preview` scripts
- Vite + React + TypeScript dependencies
- `tsconfig.json` and `tsconfig.node.json`
- `vite.config.ts`
- `index.html`

Keep the setup minimal; do not add extra tooling beyond what the first frontend shell needs.

- [ ] **Step 4: Verify the workspace files exist**

Run: `find web -maxdepth 1 -type f | sort`
Expected: PASS the checklist with the five new top-level app files present.

- [ ] **Step 5: Commit**

```bash
git add web/package.json web/tsconfig.json web/tsconfig.node.json web/vite.config.ts web/index.html
git commit -m "feat(frontend): initialize react typescript workspace"
```

## Task 2: Bootstrap the app root, router, and theme foundation

**Files:**
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/style/index.css`
- Create: `web/src/router/index.tsx`

- [ ] **Step 1: Write the failing app render test target**

Create a temporary checklist in your notes:

```text
- app bootstraps from src/main.tsx
- App.tsx renders router outlet
- router defines two routes
- global style file exists
```

- [ ] **Step 2: Verify the current src tree fails the checklist**

Run: `find web/src -maxdepth 2 -type f | sort`
Expected: FAIL because the `src` bootstrap files do not exist yet.

- [ ] **Step 3: Write the minimal app root**

Implement:

- `main.tsx` mounting React app
- `App.tsx` returning router provider or route outlet root
- `router/index.tsx` defining routes for `/` and `/system/ldap`
- `style/index.css` defining warm light theme tokens, base reset, and app-wide typography/colors

Do not add more routes than `Dashboard` and `System > LDAP`.

- [ ] **Step 4: Verify the src bootstrap files exist**

Run: `find web/src -maxdepth 2 -type f | sort`
Expected: PASS with the bootstrap, router, and global style files present.

- [ ] **Step 5: Commit**

```bash
git add web/src/main.tsx web/src/App.tsx web/src/router/index.tsx web/src/style/index.css
git commit -m "feat(frontend): bootstrap app root and router"
```

## Task 3: Build the shared app shell and navigation

**Files:**
- Create: `web/src/layout/AppShell.tsx`
- Create: `web/src/layout/AppShell.module.css`
- Create: `web/src/components/nav/SideNav.tsx`
- Create: `web/src/components/nav/SideNav.module.css`
- Modify: `web/src/router/index.tsx`

- [ ] **Step 1: Write the failing shell structure checklist**

Create a temporary checklist in your notes:

```text
- shell has only Dashboard and System top-level entries
- System section can expose LDAP as an expanded child entry
- shell styling matches the light beige productized direction
```

- [ ] **Step 2: Verify the app currently fails the shell checklist**

Run: `sed -n '1,240p' web/src/router/index.tsx`
Expected: FAIL because there is no shell wrapper or structured navigation yet.

- [ ] **Step 3: Write the minimal shell implementation**

Implement:

- `AppShell.tsx` with sidebar + main content frame
- `SideNav.tsx` rendering only:
  - `Dashboard`
  - `System`
  - child item `LDAP` under `System`
- beige/light productized styling in CSS modules
- router nesting so:
  - `/` renders inside the shell
  - `/system/ldap` renders inside the shell

- [ ] **Step 4: Verify the shell structure**

Run:

```bash
sed -n '1,260p' web/src/layout/AppShell.tsx
sed -n '1,260p' web/src/components/nav/SideNav.tsx
sed -n '1,260p' web/src/router/index.tsx
```

Expected: PASS with only the intended two top-level entries and the LDAP child route.

- [ ] **Step 5: Commit**

```bash
git add web/src/layout/AppShell.tsx web/src/layout/AppShell.module.css web/src/components/nav/SideNav.tsx web/src/components/nav/SideNav.module.css web/src/router/index.tsx
git commit -m "feat(frontend): add ops shell and minimal navigation"
```

## Task 4: Create shared page primitives for headers and empty states

**Files:**
- Create: `web/src/components/page/PageHeader.tsx`
- Create: `web/src/components/page/PageHeader.module.css`
- Create: `web/src/components/state/EmptyState.tsx`
- Create: `web/src/components/state/EmptyState.module.css`

- [ ] **Step 1: Write the failing component reuse checklist**

Create a temporary checklist in your notes:

```text
- page header component exists
- empty state component exists
- both pages can reuse the same primitives
```

- [ ] **Step 2: Verify the shared primitives do not exist yet**

Run: `find web/src/components -maxdepth 3 -type f | sort`
Expected: FAIL the checklist because no shared page primitives exist yet.

- [ ] **Step 3: Write the minimal reusable primitives**

Implement:

- `PageHeader` for title + description + optional action slot
- `EmptyState` for “coming soon / under construction” content

Keep both components presentation-only.

- [ ] **Step 4: Verify the components exist**

Run: `find web/src/components -maxdepth 3 -type f | sort`
Expected: PASS with the new page and state component files present.

- [ ] **Step 5: Commit**

```bash
git add web/src/components/page/PageHeader.tsx web/src/components/page/PageHeader.module.css web/src/components/state/EmptyState.tsx web/src/components/state/EmptyState.module.css
git commit -m "feat(frontend): add shared page primitives"
```

## Task 5: Build the Dashboard empty-state page

**Files:**
- Create: `web/src/view/dashboard/DashboardPage.tsx`
- Create: `web/src/view/dashboard/DashboardPage.module.css`

- [ ] **Step 1: Write the failing dashboard checklist**

Create a temporary checklist in your notes:

```text
- dashboard is not a fake metrics homepage
- dashboard clearly reads as under construction
- dashboard still looks intentional, not blank
```

- [ ] **Step 2: Verify no dashboard page exists yet**

Run: `find web/src/view/dashboard -maxdepth 2 -type f | sort`
Expected: FAIL because the dashboard page files do not exist yet.

- [ ] **Step 3: Write the dashboard page**

Implement a page that contains:

- shared page header
- “coming soon / under construction” message
- short explanation that the dashboard is reserved for future platform overview
- light productized styling consistent with the shell

Do not add fake charts, fake KPIs, or fake service tables.

- [ ] **Step 4: Verify the dashboard page content**

Run:

```bash
sed -n '1,240p' web/src/view/dashboard/DashboardPage.tsx
sed -n '1,240p' web/src/view/dashboard/DashboardPage.module.css
```

Expected: PASS with a clear empty-state dashboard and no fake operational data.

- [ ] **Step 5: Commit**

```bash
git add web/src/view/dashboard/DashboardPage.tsx web/src/view/dashboard/DashboardPage.module.css
git commit -m "feat(frontend): add dashboard empty state"
```

## Task 6: Build the System > LDAP static shell page

**Files:**
- Create: `web/src/view/system/SystemLdapPage.tsx`
- Create: `web/src/view/system/SystemLdapPage.module.css`

- [ ] **Step 1: Write the failing LDAP page checklist**

Create a temporary checklist in your notes:

```text
- page is clearly under System > LDAP
- page looks like a static configuration shell
- page includes connection settings area
- page includes sync/status area
- page does not imply real LDAP logic is implemented
```

- [ ] **Step 2: Verify the LDAP page does not exist yet**

Run: `find web/src/view/system -maxdepth 2 -type f | sort`
Expected: FAIL because the LDAP page files do not exist yet.

- [ ] **Step 3: Write the LDAP page**

Implement a static page containing:

- shared page header with `System / LDAP`
- connection settings section
- sync/status section
- placeholder action such as “Test Connection”
- clear copy that the current phase focuses on display structure, not real integration

Keep the content static and presentation-focused.

- [ ] **Step 4: Verify the LDAP page structure**

Run:

```bash
sed -n '1,260p' web/src/view/system/SystemLdapPage.tsx
sed -n '1,260p' web/src/view/system/SystemLdapPage.module.css
```

Expected: PASS with the required sections and no real logic.

- [ ] **Step 5: Commit**

```bash
git add web/src/view/system/SystemLdapPage.tsx web/src/view/system/SystemLdapPage.module.css
git commit -m "feat(frontend): add system ldap shell page"
```

## Task 7: Verify buildability and optionally document startup

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Install frontend dependencies**

Run: `cd web && npm install`
Expected: dependencies install successfully.

- [ ] **Step 2: Verify the frontend builds**

Run: `cd web && npm run build`
Expected: PASS

- [ ] **Step 3: Verify the dev shell starts**

Run: `cd web && npm run dev`
Expected: Vite dev server starts successfully.

- [ ] **Step 4: Add a short README note only if the frontend is runnable**

If build and dev startup succeed, add a short frontend startup note to `README.md` covering:

- frontend directory
- install command
- dev command

Do not expand README beyond that minimal note.

- [ ] **Step 5: Commit**

```bash
git add README.md web/package.json web/tsconfig.json web/tsconfig.node.json web/vite.config.ts web/index.html web/src
git commit -m "feat(frontend): add ops shell and ldap static frontend"
```
