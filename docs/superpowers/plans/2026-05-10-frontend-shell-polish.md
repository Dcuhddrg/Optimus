# Frontend Shell Polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework the authenticated front-end shell so the `System` navigation becomes a natural secondary section and the sidebar, topbar, and page chrome share one lighter visual language.

**Architecture:** Keep all routing, auth, and permission behavior intact while tightening the shell structure around the existing React Router layout. Most work stays inside `SideNav`, `AppShell`, and shared shell CSS tokens; page-level changes are limited to small header and container adjustments where the old chrome clashes with the new shell.

**Tech Stack:** React 18, React Router 6, TypeScript, Vite 5, CSS Modules, global CSS variables, `npm run build` for verification.

---

## Scope Notes

- This plan is limited to front-end shell structure and styling for authenticated pages.
- Do not change backend APIs, auth state, permission visibility logic, or route definitions.
- Keep `.superpowers/brainstorm/**` untracked artifacts untouched.
- The repo does not include a dedicated front-end test runner. Use `npm run build` plus explicit manual route checks instead of adding new test dependencies for this polish task.
- Preserve the current `System` auto-expand behavior on `/system/*` routes while keeping manual collapse/expand available on non-`/system` routes.

## File Map

### Modify

- `web/src/components/nav/SideNav.tsx` — simplify the nav structure so primary items and the `System` section share one navigation model while preserving permission checks and expand/collapse behavior
- `web/src/components/nav/SideNav.module.css` — replace the current mixed card/accordion styling with a lighter, unified sidebar visual system
- `web/src/components/locale/LanguageSwitcher.module.css` — align the language switcher with the lighter topbar control treatment without changing locale-switching behavior
- `web/src/layout/AppShell.tsx` — lightly regroup topbar content without introducing route metadata or changing layout ownership
- `web/src/layout/AppShell.module.css` — make the shell grid, topbar, session controls, and main content wrapper share the new chrome language
- `web/src/style/index.css` — tighten global shell tokens for background, surface, borders, text, and shared page-card behavior
- `web/src/components/page/PageHeader.module.css` — align shared page-header spacing and weight with the new shell where pages already use `PageHeader`
- `web/src/view/system/SystemUsersPage.module.css` — small container/table row adjustments only if the old page chrome clashes with the new shell
- `web/src/view/system/SystemRolesPage.module.css` — small card spacing and border adjustments only if needed for shell consistency
- `web/src/view/system/SystemLdapPage.module.css` — reduce panel weight so the LDAP form does not overpower the lighter shell
- `web/src/view/system/SystemAuditLogsPage.module.css` — small list container adjustments only if needed for shell consistency

### Do Not Modify Unless Forced by Build/Type Errors

- `web/src/router/index.tsx` — routes are already correct
- `web/src/auth/**` — auth flow is out of scope
- `web/src/view/system/*.tsx` — keep business rendering and behavior intact unless a tiny wrapper/classname adjustment is unavoidable

## Manual Verification Checklist

Use this exact checklist during implementation:

- Use an account that can see all `System` child routes; otherwise the route-by-route nav verification is incomplete
- `/` shows the refreshed shell without layout breakage
- `/system/users` auto-expands `System`, highlights the parent section, and highlights `Users`
- `/system/roles`, `/system/ldap`, and `/system/audit-logs` each highlight the correct child item
- On a non-`/system` route, clicking `System` manually collapses and expands the section
- The topbar session controls still work visually at desktop width and wrap cleanly on narrower widths
- At a narrower viewport, the sidebar column, main content, and expanded `System` child list remain readable with no overlap, clipping, or severe squeeze
- LDAP form, roles cards, users rows, and audit list do not look heavier than the surrounding shell

## Task 1: Tighten Global Shell Tokens and AppShell Chrome

**Files:**
- Modify: `web/src/style/index.css`
- Modify: `web/src/layout/AppShell.tsx`
- Modify: `web/src/layout/AppShell.module.css`
- Modify: `web/src/components/locale/LanguageSwitcher.module.css`

- [ ] **Step 1: Write down the shell invariants in `AppShell.module.css` comments before changing styles**

Add a short comment block near the shell layout describing the non-negotiables:

```css
/*
Shell invariants:
- fixed two-column desktop shell
- lighter topbar than content
- session controls wrap without overlap
*/
```

This keeps the implementation anchored on layout behavior rather than ad-hoc visual tweaks.

- [ ] **Step 2: Update the global shell tokens in `index.css`**

Tighten the existing warm palette into reusable chrome variables, for example:

```css
:root {
  --shell-bg: #f3efe7;
  --shell-surface: rgba(255, 252, 247, 0.82);
  --shell-surface-strong: rgba(255, 251, 246, 0.94);
  --shell-border: rgba(103, 84, 65, 0.12);
  --shell-border-strong: rgba(103, 84, 65, 0.18);
  --shell-muted: #6d6256;
  --shell-radius: 22px;
}
```

Keep the variables limited to shell/chrome concerns. Do not turn this into a broad theme rewrite.

- [ ] **Step 3: Simplify the `AppShell` markup only where it improves grouping**

Keep the same layout ownership, but allow a small wrapper split such as:

```tsx
<header className={styles.topbar}>
  <div className={styles.topbarCopy}>...</div>
  <div className={styles.topbarActions}>...</div>
</header>
```

Do not introduce route-title lookup, extra providers, or new shell state.

- [ ] **Step 4: Restyle the shell grid, topbar, and main content wrapper**

Make the authenticated shell read as one continuous chrome:

- sidebar and content share one background family
- topbar becomes thinner and calmer
- session controls use consistent border/background treatment
- language switcher matches the same control family as the user badge and sign-out action
- main content gets a clearer inner canvas with controlled radius and border

Keep the existing desktop grid and current responsive breakpoint unless a minimal breakpoint tweak is required.

- [ ] **Step 5: Run the front-end build**

Run: `cd web && npm run build`

Expected: PASS. Failures at this point should only be CSS-module name mismatches, JSX wrapper mistakes, or type/import mistakes.

- [ ] **Step 6: Commit**

```bash
git add web/src/style/index.css web/src/layout/AppShell.tsx web/src/layout/AppShell.module.css web/src/components/locale/LanguageSwitcher.module.css
git commit -m "Unify the authenticated shell chrome" \
  -m "Reduce the visual weight of the app shell so the sidebar, topbar, and main canvas read as one surface family without changing routing or auth behavior." \
  -m "Constraint: Must preserve the current app-shell structure and responsive wrapping behavior" \
  -m "Rejected: Add route metadata for dynamic topbar titles | outside the approved shell-polish scope" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Keep shell tokens scoped to chrome concerns; do not expand this into a new theme system" \
  -m "Tested: cd web && npm run build"
```

## Task 2: Rework Sidebar Structure and the System Section

**Files:**
- Modify: `web/src/components/nav/SideNav.tsx`
- Modify: `web/src/components/nav/SideNav.module.css`

- [ ] **Step 1: Add a small nav item definition inside `SideNav.tsx` before changing markup**

Create tiny local data definitions so the structure reads as one system:

```tsx
const systemItems = [
  showUsers && { to: '/system/users', label: t('shell.users') },
  showRoles && { to: '/system/roles', label: t('shell.roles') },
  showLDAP && { to: '/system/ldap', label: t('shell.ldap') },
  showAudit && { to: '/system/audit-logs', label: t('shell.auditLogs') },
].filter((item): item is { to: string; label: string } => Boolean(item));
```

Keep this local to the file. Do not extract a new nav abstraction for one sidebar.

- [ ] **Step 2: Restructure the markup so primary nav and the `System` section share one rhythm**

Aim for a flow like:

```tsx
<nav className={styles.nav}>
  <NavLink ...>Dashboard</NavLink>
  <section className={styles.section}>
    <button ...>System</button>
    {expanded ? <div className={styles.sectionChildren}>...</div> : null}
  </section>
</nav>
```

Requirements:

- keep current `useEffect` auto-expand on `/system/*`
- keep manual toggle on non-system routes
- do not change permission gating logic

- [ ] **Step 3: Replace the current heavy accordion styling with a unified sidebar language**

Implement:

- lighter brand block
- matched visual treatment for `Dashboard` and `System`
- `System` children inside a subtle inset panel or guided rail
- calmer active and hover states with less shadow
- clearer parent-active + child-active relationship on system routes

Prefer a small number of well-named CSS classes over duplicated active/idle variants when composition can keep the file simpler.

- [ ] **Step 4: Run the front-end build**

Run: `cd web && npm run build`

Expected: PASS.

- [ ] **Step 5: Manually verify nav behavior in the browser**

Check:

- `/` allows manual `System` collapse/expand
- `/system/users` forces the section open
- active styling does not make the child list look like a detached card

If behavior is wrong, fix `SideNav.tsx` before moving on.

- [ ] **Step 6: Commit**

```bash
git add web/src/components/nav/SideNav.tsx web/src/components/nav/SideNav.module.css
git commit -m "Make the system nav feel native to the sidebar" \
  -m "Restructure the sidebar so the system links read as a natural secondary section instead of a separate add-on block while preserving permission and expand-state behavior." \
  -m "Constraint: Must preserve current menu visibility rules and auto-expand on system routes" \
  -m "Rejected: Split the sidebar into reusable nav primitives | unnecessary abstraction for a single shell component" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Keep the system child container visually nested, never floating as an independent card" \
  -m "Tested: cd web && npm run build; manual sidebar collapse/expand and system-route active-state check"
```

## Task 3: Align Shared Page Header and System Page Containers

**Files:**
- Modify: `web/src/components/page/PageHeader.module.css`
- Modify: `web/src/view/system/SystemUsersPage.module.css`
- Modify: `web/src/view/system/SystemRolesPage.module.css`
- Modify: `web/src/view/system/SystemLdapPage.module.css`
- Modify: `web/src/view/system/SystemAuditLogsPage.module.css`

- [ ] **Step 1: Normalize the shared page-header tone**

Adjust `PageHeader.module.css` so pages already using `PageHeader` inherit the lighter shell:

- slightly tighter eyebrow spacing
- calmer title size/weight relationship
- actions align with the thinner topbar language

Do not change the `PageHeader` component API unless a class-level limitation forces it.

- [ ] **Step 2: Soften the heaviest system-page containers**

Apply only targeted CSS adjustments:

- users rows: slightly calmer borders/background and spacing
- roles cards: align card surface with shell surface
- LDAP panel/buttons: reduce gradient and shadow weight
- audit items: align with the new list/card tone

Avoid renaming class structures or touching page logic. Keep these as shell-coordination adjustments, not page redesigns.

- [ ] **Step 3: Run the front-end build**

Run: `cd web && npm run build`

Expected: PASS.

- [ ] **Step 4: Manually verify the four system pages**

Open:

- `/system/users`
- `/system/roles`
- `/system/ldap`
- `/system/audit-logs`

Confirm each page sits comfortably inside the new shell and none of the inner panels overpower the chrome.

- [ ] **Step 5: Commit**

```bash
git add web/src/components/page/PageHeader.module.css \
  web/src/view/system/SystemUsersPage.module.css \
  web/src/view/system/SystemRolesPage.module.css \
  web/src/view/system/SystemLdapPage.module.css \
  web/src/view/system/SystemAuditLogsPage.module.css
git commit -m "Bring system page surfaces in line with the new shell" \
  -m "Trim the heaviest page-level borders, gradients, and spacing so existing system screens sit naturally inside the lighter shell without changing their logic." \
  -m "Constraint: Must keep page-level work limited to shell-coordination CSS changes" \
  -m "Rejected: Rebuild system pages around new content components | outside the approved polish scope" \
  -m "Confidence: medium" \
  -m "Scope-risk: narrow" \
  -m "Directive: If a page still feels off, prefer small surface-tone adjustments over structural rewrites" \
  -m "Tested: cd web && npm run build; manual system-page shell-coordination check"
```

## Task 4: Final Verification and Cleanup

**Files:**
- Review only: `web/src/components/nav/SideNav.tsx`
- Review only: `web/src/components/nav/SideNav.module.css`
- Review only: `web/src/layout/AppShell.tsx`
- Review only: `web/src/layout/AppShell.module.css`
- Review only: `web/src/style/index.css`
- Review only: `web/src/components/page/PageHeader.module.css`
- Review only: `web/src/view/system/SystemUsersPage.module.css`
- Review only: `web/src/view/system/SystemRolesPage.module.css`
- Review only: `web/src/view/system/SystemLdapPage.module.css`
- Review only: `web/src/view/system/SystemAuditLogsPage.module.css`

- [ ] **Step 1: Remove any dead selectors or temporary comments added during styling**

Clean out:

- stale duplicated active/idle classes no longer referenced
- exploratory comments that are no longer needed
- unused wrappers introduced and then abandoned

Keep the code easier to reason about than before the polish started.

- [ ] **Step 2: Run the full front-end build one last time**

Run: `cd web && npm run build`

Expected: PASS.

- [ ] **Step 3: Execute the full manual verification checklist**

Verify:

- dashboard shell
- all system-route active states
- manual collapse/expand on non-system routes
- topbar wrapping on narrower widths
- expanded sidebar readability on narrower widths
- system pages no longer visually overpower the shell

Record any residual issue before claiming completion.

- [ ] **Step 4: Commit the final polish pass**

```bash
git add web/src/components/nav/SideNav.tsx \
  web/src/components/nav/SideNav.module.css \
  web/src/layout/AppShell.tsx \
  web/src/layout/AppShell.module.css \
  web/src/style/index.css \
  web/src/components/page/PageHeader.module.css \
  web/src/view/system/SystemUsersPage.module.css \
  web/src/view/system/SystemRolesPage.module.css \
  web/src/view/system/SystemLdapPage.module.css \
  web/src/view/system/SystemAuditLogsPage.module.css
git commit -m "Finish the authenticated shell polish" \
  -m "Finalize the lighter shell presentation and confirm the system navigation, topbar, and page surfaces now read as one coherent chrome layer." \
  -m "Constraint: Must ship without changing application behavior or adding new frontend dependencies" \
  -m "Rejected: Leave page-surface inconsistencies for later | would undermine the shell-level polish immediately visible to users" \
  -m "Confidence: medium" \
  -m "Scope-risk: moderate" \
  -m "Directive: Future shell work should extend the shared surface language before inventing page-specific chrome" \
  -m "Tested: cd web && npm run build; manual dashboard/system shell verification"
```

## Implementation Notes

- If a tiny `.tsx` change is required inside a system page to add a wrapper class for shell coordination, keep it as a single-purpose change and note it in the commit message for that task.
- Do not introduce icons, new fonts, animation libraries, or decorative flourishes not covered by the approved spec.
- If the current responsive layout reveals a hard limitation that cannot be solved in CSS alone, stop and reassess before introducing new shell state.
