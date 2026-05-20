# Frontend Language Switch Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a front-end-only Chinese/English language switch for the existing Optimus shell, login flow, and system pages described in `docs/superpowers/specs/2026-05-08-frontend-language-switch-design.md`.

**Architecture:** Add a lightweight locale layer inside `web/src` with a shared provider, static dictionaries, and small label helpers for stable `code` and enum values. Keep all backend contracts unchanged; pages only consume `t(...)` and label helper functions, while login and shell share one persisted locale state.

**Tech Stack:** React 18, React Router 6, TypeScript, Vite 5, CSS Modules, browser `localStorage`, browser `navigator.language`.

---

## Scope Notes

- This plan is limited to front-end display behavior. Do not change backend APIs, database models, or Go services.
- Do not add an internationalization dependency such as `react-i18next`.
- Keep `.superpowers/brainstorm/**` untracked artifacts untouched.
- The repo does not currently include a front-end test runner. Favor pure helper extraction, `npm run build`, and manual browser verification over adding test dependencies for this feature.

## File Map

### Create

- `web/src/locale/types.ts` — locale type definitions and supported locale list
- `web/src/locale/storage.ts` — `optimus.locale` read/write helpers
- `web/src/locale/resolveLocale.ts` — browser-language normalization and fallback logic
- `web/src/locale/messages.ts` — `en` and `zh-CN` fixed-copy dictionaries
- `web/src/locale/systemLabels.ts` — enum and `code` label mapping helpers
- `web/src/locale/LanguageProvider.tsx` — provider, context, and locale switching state
- `web/src/locale/useI18n.ts` — typed hook for `locale`, `setLocale`, and `t`
- `web/src/components/locale/LanguageSwitcher.tsx` — reusable two-language switcher
- `web/src/components/locale/LanguageSwitcher.module.css` — switcher styling consistent with the current shell

### Modify

- `web/src/router/index.tsx` — wrap the app in the locale provider and route the root dashboard through the dedicated dashboard page
- `web/src/App.tsx` — only if provider composition is cleaner at the app root
- `web/src/layout/AppShell.tsx` — translate shell copy and mount the topbar switcher
- `web/src/components/nav/SideNav.tsx` — translate navigation labels and aria copy
- `web/src/view/auth/LoginPage.tsx` — translate login copy, mount login-page switcher, and localize default front-end error text
- `web/src/view/dashboard/DashboardPage.tsx` — translate dashboard placeholder copy and make it the routed dashboard surface
- `web/src/view/system/ForbiddenPage.tsx` — translate access-denied copy
- `web/src/view/system/SystemUsersPage.tsx` — translate page copy and map user `status`, `source`, and `roles`
- `web/src/view/system/SystemRolesPage.tsx` — translate page copy and map role `code`, `status`, and permission labels used by the page
- `web/src/view/system/SystemLdapPage.tsx` — translate fixed form copy while leaving backend-returned raw messages untouched
- `web/src/view/system/SystemAuditLogsPage.tsx` — translate page copy and map stable audit result/status labels where possible
- `web/src/components/page/PageHeader.tsx` — only if a small prop or helper adjustment is needed for cleaner localized headers

## Manual Verification Checklist

Use this exact checklist during implementation:

- Browser language `zh-*` with no saved locale picks `zh-CN`
- Browser language other than `zh-*` with no saved locale picks `en`
- Manual switch writes `optimus.locale` in `localStorage`
- Refresh keeps the chosen language
- Login page and AppShell switch independently visible but share the same state
- System page fixed copy changes with locale
- Unknown `code` or enum values render the original raw value instead of blank text

## Task 1: Build the Locale Core

**Files:**
- Create: `web/src/locale/types.ts`
- Create: `web/src/locale/storage.ts`
- Create: `web/src/locale/resolveLocale.ts`
- Create: `web/src/locale/LanguageProvider.tsx`
- Create: `web/src/locale/useI18n.ts`
- Modify: `web/src/router/index.tsx`
- Modify: `web/src/App.tsx` if needed

- [ ] **Step 1: Write a tiny behavior table in `resolveLocale.ts` comments before implementation**

Capture the expected normalization rules next to the helper skeleton:

```ts
// 'zh-CN' -> 'zh-CN'
// 'zh-TW' -> 'zh-CN'
// 'en-US' -> 'en'
// undefined -> 'en'
```

This is the feature's executable contract until the repo grows a front-end test runner.

- [ ] **Step 2: Add the locale primitives**

Define:

```ts
export type AppLocale = 'zh-CN' | 'en';
export const SUPPORTED_LOCALES = ['zh-CN', 'en'] as const;
export const LOCALE_STORAGE_KEY = 'optimus.locale';
```

Add helpers:

```ts
export function resolveLocale(input?: string | null): AppLocale;
export function getStoredLocale(): AppLocale | null;
export function setStoredLocale(locale: AppLocale): void;
```

- [ ] **Step 3: Add the provider and hook**

Implement a provider that:

- reads the stored locale once on mount
- falls back to `navigator.language`
- exposes `locale`, `setLocale`, and `t`
- persists locale changes to `localStorage`

Keep the context value small:

```ts
interface I18nContextValue {
  locale: AppLocale;
  setLocale: (locale: AppLocale) => void;
  t: (key: TranslationKey) => string;
}
```

- [ ] **Step 4: Wire the provider into the root route tree**

Wrap the existing auth/router root so both `/login` and protected pages share one locale state. Prefer changing `web/src/router/index.tsx` first; only touch `web/src/App.tsx` if that makes the tree clearer.

- [ ] **Step 5: Run the front-end build**

Run: `cd web && npm run build`

Expected: PASS. Failures at this stage should only be missing imports, bad types, or incomplete provider wiring.

- [ ] **Step 6: Commit**

```bash
git add web/src/locale web/src/router/index.tsx web/src/App.tsx
git commit -m "Add a shared front-end locale state" \
  -m "Constraint: Must support login and authenticated pages without changing backend APIs" \
  -m "Rejected: Keep per-page locale state | would desynchronize login and shell switching" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Keep locale detection and storage helpers pure so later tests can target them directly" \
  -m "Tested: cd web && npm run build"
```

## Task 2: Add Dictionaries and System Label Helpers

**Files:**
- Create: `web/src/locale/messages.ts`
- Create: `web/src/locale/systemLabels.ts`
- Modify: `web/src/types/system.ts` only if helper typing needs a small shared alias

- [ ] **Step 1: Add fixed-copy dictionaries**

Create `en` and `zh-CN` dictionaries for these groups:

- `common`
- `auth`
- `shell`
- `dashboard`
- `system.forbidden`
- `system.users`
- `system.roles`
- `system.ldap`
- `system.audit`

Keep keys stable and semantic, for example:

```ts
auth.signInTitle
auth.usernameLabel
shell.signOut
system.users.emptyRoles
system.ldap.testConnection
```

- [ ] **Step 2: Add label helpers for enums and stable codes**

Implement helpers such as:

```ts
export function getStatusLabel(value: string, locale: AppLocale): string;
export function getUserSourceLabel(value: string, locale: AppLocale): string;
export function getRoleCodeLabel(code: string, locale: AppLocale): string;
export function getPermissionCodeLabel(code: string, locale: AppLocale): string;
export function getPermissionTypeLabel(type: string, locale: AppLocale): string;
```

Map known values, but always fall back to the original raw value.

- [ ] **Step 3: Keep translation concerns separate**

Do not mix page-level copy and code-label tables into one object. `messages.ts` owns fixed copy; `systemLabels.ts` owns enum/code mapping.

- [ ] **Step 4: Run the front-end build**

Run: `cd web && npm run build`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/src/locale/messages.ts web/src/locale/systemLabels.ts web/src/types/system.ts
git commit -m "Define front-end dictionaries and code labels" \
  -m "Constraint: System screens need localized labels without backend schema changes" \
  -m "Rejected: Use backend display_name fields for translation | current APIs do not provide locale-specific values" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Unknown code and enum values must always fall back to the original raw value" \
  -m "Tested: cd web && npm run build"
```

## Task 3: Add the Shared Language Switcher and Translate the Shell

**Files:**
- Create: `web/src/components/locale/LanguageSwitcher.tsx`
- Create: `web/src/components/locale/LanguageSwitcher.module.css`
- Modify: `web/src/view/auth/LoginPage.tsx`
- Modify: `web/src/layout/AppShell.tsx`
- Modify: `web/src/components/nav/SideNav.tsx`
- Modify: `web/src/router/index.tsx`
- Modify: `web/src/view/dashboard/DashboardPage.tsx`
- Modify: `web/src/view/system/ForbiddenPage.tsx`

- [ ] **Step 1: Build the switcher component**

Implement one reusable switcher that reads `locale` and `setLocale()` from `useI18n()`. Keep the UI simple: two explicit choices, no dropdown.

- [ ] **Step 2: Mount the switcher on the login page**

Translate:

- eyebrow
- title
- description
- username/password labels
- submit button
- front-end fallback error text (`Login failed`)

Place the switcher in the login-page top-right area without changing the form flow.

- [ ] **Step 3: Mount the switcher in the authenticated shell**

Translate:

- shell eyebrow
- shell title
- sign-out button
- nav labels
- nav aria label
- dashboard / forbidden placeholder copy

Keep the existing auth and logout behavior untouched.

- [ ] **Step 4: Run the front-end build**

Run: `cd web && npm run build`

Expected: PASS.

- [ ] **Step 5: Run the app locally and verify switcher behavior**

Run: `cd web && npm run dev -- --host 127.0.0.1`

Manual checks:

- open `/login`
- switch between `中文` and `English`
- refresh and confirm persistence
- log in and confirm the topbar switcher matches the login-page choice

- [ ] **Step 6: Commit**

```bash
git add web/src/components/locale web/src/view/auth/LoginPage.tsx web/src/layout/AppShell.tsx web/src/components/nav/SideNav.tsx web/src/router/index.tsx web/src/view/dashboard/DashboardPage.tsx web/src/view/system/ForbiddenPage.tsx
git commit -m "Expose language switching across the shell" \
  -m "Constraint: Users must be able to switch language before and after authentication" \
  -m "Rejected: Put the control only in authenticated pages | first-time login would still start in the wrong language with no immediate override" \
  -m "Confidence: high" \
  -m "Scope-risk: moderate" \
  -m "Directive: Keep the switcher visually lightweight and reuse one component in both login and shell contexts" \
  -m "Tested: cd web && npm run build" \
  -m "Tested: manual login-page and shell language switching"
```

## Task 4: Translate System Pages and Apply Code/Enum Mapping

**Files:**
- Modify: `web/src/view/system/SystemUsersPage.tsx`
- Modify: `web/src/view/system/SystemRolesPage.tsx`
- Modify: `web/src/view/system/SystemLdapPage.tsx`
- Modify: `web/src/view/system/SystemAuditLogsPage.tsx`
- Modify: `web/src/components/page/PageHeader.tsx` if needed

- [ ] **Step 1: Translate the Users page**

Replace fixed copy with dictionary lookups and map:

- `user.source`
- `user.status`
- `user.roles`

Use a localized empty fallback instead of the hardcoded `No roles`.

- [ ] **Step 2: Translate the Roles page**

Replace fixed copy with dictionary lookups and map:

- role `code`
- role `status`
- permission-related action labels

If the page later surfaces permission codes directly, route them through `getPermissionCodeLabel(...)`.

- [ ] **Step 3: Translate the LDAP page**

Replace all form labels, button labels, page titles, and front-end fallback error strings with `t(...)`.

Do not rewrite backend-returned `result.message` strings. Display them as-is.

- [ ] **Step 4: Translate the Audit Logs page**

Translate page header copy and map stable fields where possible:

- `result`
- known `resourceType` values if they are stable enough

Leave dynamic `action` values raw unless the backend action vocabulary is already stable and worth mapping in this same pass.

- [ ] **Step 5: Run the front-end build**

Run: `cd web && npm run build`

Expected: PASS.

- [ ] **Step 6: Verify code/enum fallback manually**

Use the browser with realistic API responses and confirm:

- mapped status/source/code values show translated labels
- unknown values still render the original raw string

- [ ] **Step 7: Commit**

```bash
git add web/src/view/system/SystemUsersPage.tsx web/src/view/system/SystemRolesPage.tsx web/src/view/system/SystemLdapPage.tsx web/src/view/system/SystemAuditLogsPage.tsx web/src/components/page/PageHeader.tsx
git commit -m "Localize system management pages" \
  -m "Constraint: Front-end must present bilingual system copy without changing backend response shapes" \
  -m "Rejected: Translate backend-originated free text in the client | message semantics are not stable enough for safe client-side remapping" \
  -m "Confidence: medium" \
  -m "Scope-risk: moderate" \
  -m "Directive: Keep fixed-copy translation and raw backend-message display as separate paths" \
  -m "Tested: cd web && npm run build" \
  -m "Tested: manual system-page locale checks"
```

## Task 5: Final Verification and Cleanup

**Files:**
- Modify: any touched front-end files only if verification reveals a defect

- [ ] **Step 1: Check formatting and type/build health**

Run: `cd web && npm run build`

Expected: PASS.

- [ ] **Step 2: Verify the locale persistence behavior from a clean browser state**

Manual checks:

- clear `optimus.locale`
- confirm browser-language defaulting
- switch to the other locale
- refresh
- confirm the stored value wins

- [ ] **Step 3: Verify authenticated-page behavior**

Manual checks after login:

- topbar and navigation are translated
- at least one system page shows localized fixed copy
- at least one code/enum-derived label is translated
- logout still redirects to `/login`

- [ ] **Step 4: Check the git diff for accidental spread**

Run:

```bash
git diff --check
git status --short
```

Expected: no whitespace errors, only intended front-end file changes.

- [ ] **Step 5: Commit the final fixups if needed**

If verification required any cleanup edits, commit them with a Lore-format message that names the concrete verification issue fixed. If no cleanup edits were needed, do not create an extra commit.
