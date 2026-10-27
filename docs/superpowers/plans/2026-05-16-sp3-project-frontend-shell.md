# SP3 Project Frontend Shell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the frontend project management surface, request-scoped project context, topbar project switcher, and three-section navigation required for SP3.

**Architecture:** Keep SP3 focused on frontend project context and shell integration. Add typed `project` API functions over the SP1 backend, store the selected project code in `localStorage`, expose it through `ProjectProvider`, and have `apiRequest` add `X-Project-Code` whenever a current project exists. Add `/projects` and `/projects/:code/members` pages, mount `ProjectSwitcher` in `AppShell`, and refactor `SideNav` into Overview / Current Project / System sections without implementing SP4 cloud pages yet.

**Tech Stack:** React 18, React Router 6, TypeScript, CSS modules, Vite.

**Spec reference:** `docs/superpowers/specs/2026-05-14-multi-project-cloud-isolation-design.md` §Frontend UX and §SP3. Backend dependencies SP1 + SP2 are already merged into `main`.

---

## Requirements Summary

- Add typed frontend support for project CRUD and project membership endpoints.
- Track the current project globally and persist the selected `project.code`.
- Inject `X-Project-Code` into all API requests after a current project is selected.
- Add a topbar `ProjectSwitcher` that lists accessible projects and links to `/projects`.
- Add project list and project member pages.
- Add `/projects` and `/projects/:code/members` routes.
- Refactor `SideNav` into Overview, Current Project, and System sections.
- Keep SP4 out of scope: no cloud account/asset page implementation in this plan.

## Acceptance Criteria

- After login/current-user load, `ProjectProvider` fetches accessible projects and chooses the stored project if still accessible, otherwise the first accessible project.
- `apiRequest` sends `X-Project-Code: <current.code>` when a project is selected and omits it otherwise.
- `ProjectSwitcher` appears in `AppShell` for authenticated users, displays the current project, lets the user switch projects, and has a Manage Projects link.
- `/projects` lists projects, supports create/update/delete with existing backend validation surfaced as errors.
- `/projects/:code/members` lists members, supports add/update/remove role actions.
- `SideNav` shows Overview, Current Project, and System sections; project management is gated by `project.menu`; cloud links are gated by cloud menu permissions but may point to placeholder routes until SP4.
- `npm run build` passes.

## File Structure

**Create:**
- `web/src/types/project.ts` — project and membership DTO types.
- `web/src/api/project.ts` — typed project API functions.
- `web/src/auth/ProjectContext.tsx` — project list/current project provider and localStorage persistence.
- `web/src/components/project/ProjectSwitcher.tsx`
- `web/src/components/project/ProjectSwitcher.module.css`
- `web/src/view/project/ProjectsPage.tsx`
- `web/src/view/project/ProjectsPage.module.css`
- `web/src/view/project/ProjectMembersPage.tsx`
- `web/src/view/project/ProjectMembersPage.module.css`

**Modify:**
- `web/src/api/client.ts` — support `X-Project-Code`.
- `web/src/router/index.tsx` — mount `ProjectProvider` under `RequireAuth`; add project routes.
- `web/src/layout/AppShell.tsx`
- `web/src/layout/AppShell.module.css`
- `web/src/components/nav/SideNav.tsx`
- `web/src/components/nav/SideNav.module.css`
- `web/src/locale/messages.ts`

---

### Task 1: Project API Types and Client Header Injection

**Files:**
- Create: `web/src/types/project.ts`
- Create: `web/src/api/project.ts`
- Modify: `web/src/api/client.ts`

- [ ] **Step 1.1: Add project DTO types**

Create `web/src/types/project.ts`:

```ts
export type ProjectRole = 'owner' | 'member' | 'viewer';

export interface Project {
  id: number;
  name: string;
  code: string;
  description: string;
  ownerId: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectMember {
  userId: number;
  username: string;
  displayName: string;
  projectRole: ProjectRole | string;
}

export interface CreateProjectPayload {
  name: string;
  code: string;
  description: string;
}

export interface UpdateProjectPayload {
  name?: string;
  description?: string;
}

export interface AddProjectMemberPayload {
  userId: number;
  projectRole: ProjectRole;
}

export interface UpdateProjectMemberRolePayload {
  projectRole: ProjectRole;
}
```

- [ ] **Step 1.2: Add typed project API functions**

Create `web/src/api/project.ts`:

```ts
import { apiRequest } from './client';
import type {
  AddProjectMemberPayload,
  CreateProjectPayload,
  Project,
  ProjectMember,
  UpdateProjectMemberRolePayload,
  UpdateProjectPayload,
} from '../types/project';
import type { ListResponse } from '../types/system';

export function listProjects() {
  return apiRequest<ListResponse<Project>>('/projects');
}

export function createProject(payload: CreateProjectPayload) {
  return apiRequest<Project>('/projects', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateProject(code: string, payload: UpdateProjectPayload) {
  return apiRequest<Project>(`/projects/${code}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function deleteProject(code: string) {
  return apiRequest<{ message: string }>(`/projects/${code}`, { method: 'DELETE' });
}

export function listProjectMembers(code: string) {
  return apiRequest<ListResponse<ProjectMember>>(`/projects/${code}/members`);
}

export function addProjectMember(code: string, payload: AddProjectMemberPayload) {
  return apiRequest<{ message: string }>(`/projects/${code}/members`, {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function updateProjectMemberRole(code: string, userId: number, payload: UpdateProjectMemberRolePayload) {
  return apiRequest<{ message: string }>(`/projects/${code}/members/${userId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
}

export function removeProjectMember(code: string, userId: number) {
  return apiRequest<{ message: string }>(`/projects/${code}/members/${userId}`, {
    method: 'DELETE',
  });
}
```

- [ ] **Step 1.3: Add current project header support**

Modify `web/src/api/client.ts`:

```ts
const TOKEN_KEY = 'optimus.token';
const PROJECT_CODE_KEY = 'optimus.currentProjectCode';

let onUnauthorized: (() => void) | undefined;
let currentProjectCode: string | null = window.localStorage.getItem(PROJECT_CODE_KEY);

export function getStoredProjectCode() {
  return window.localStorage.getItem(PROJECT_CODE_KEY);
}

export function setStoredProjectCode(code: string | null) {
  currentProjectCode = code;
  if (code) {
    window.localStorage.setItem(PROJECT_CODE_KEY, code);
  } else {
    window.localStorage.removeItem(PROJECT_CODE_KEY);
  }
}
```

Then in `apiRequest`, after Authorization header:

```ts
if (currentProjectCode) {
  headers.set('X-Project-Code', currentProjectCode);
}
```

Keep existing token behavior unchanged.

- [ ] **Step 1.4: Verify**

Run: `cd /home/lilin/trae-project/Optimus/web && npm run build`
Expected: PASS.

- [ ] **Step 1.5: Commit**

```bash
git add web/src/types/project.ts web/src/api/project.ts web/src/api/client.ts
git commit -m "Add frontend project API client"
```

---

### Task 2: ProjectContext Provider

**Files:**
- Create: `web/src/auth/ProjectContext.tsx`
- Modify: `web/src/router/index.tsx`

- [ ] **Step 2.1: Add ProjectContext**

Create `web/src/auth/ProjectContext.tsx`:

```tsx
import { createContext, ReactNode, useContext, useEffect, useMemo, useState } from 'react';

import { getStoredProjectCode, setStoredProjectCode } from '../api/client';
import { listProjects } from '../api/project';
import type { Project } from '../types/project';

interface ProjectContextValue {
  projects: Project[];
  currentProject: Project | null;
  loading: boolean;
  error: string | null;
  selectProject: (code: string) => void;
  reloadProjects: () => Promise<void>;
}

const ProjectContext = createContext<ProjectContextValue | null>(null);

export function ProjectProvider({ children }: { children: ReactNode }) {
  const [projects, setProjects] = useState<Project[]>([]);
  const [currentCode, setCurrentCode] = useState<string | null>(() => getStoredProjectCode());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const currentProject = useMemo(
    () => projects.find((project) => project.code === currentCode) ?? null,
    [currentCode, projects],
  );

  async function reloadProjects() {
    setLoading(true);
    setError(null);
    try {
      const result = await listProjects();
      setProjects(result.items);
      const stored = getStoredProjectCode();
      const next = result.items.find((project) => project.code === stored) ?? result.items[0] ?? null;
      setCurrentCode(next?.code ?? null);
      setStoredProjectCode(next?.code ?? null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void reloadProjects();
  }, []);

  function selectProject(code: string) {
    setCurrentCode(code);
    setStoredProjectCode(code);
  }

  return (
    <ProjectContext.Provider value={{ projects, currentProject, loading, error, selectProject, reloadProjects }}>
      {children}
    </ProjectContext.Provider>
  );
}

export function useProject() {
  const value = useContext(ProjectContext);
  if (!value) {
    throw new Error('useProject must be used inside ProjectProvider');
  }
  return value;
}
```

- [ ] **Step 2.2: Mount ProjectProvider**

Modify `web/src/router/index.tsx`: wrap `<AppShell />` with `ProjectProvider` under `RequireAuth`:

```tsx
import { ProjectProvider } from '../auth/ProjectContext';
```

```tsx
element: (
  <ProjectProvider>
    <AppShell />
  </ProjectProvider>
),
```

- [ ] **Step 2.3: Verify**

Run: `cd /home/lilin/trae-project/Optimus/web && npm run build`
Expected: PASS.

- [ ] **Step 2.4: Commit**

```bash
git add web/src/auth/ProjectContext.tsx web/src/router/index.tsx
git commit -m "Add frontend project context"
```

---

### Task 3: ProjectSwitcher in AppShell

**Files:**
- Create: `web/src/components/project/ProjectSwitcher.tsx`
- Create: `web/src/components/project/ProjectSwitcher.module.css`
- Modify: `web/src/layout/AppShell.tsx`
- Modify: `web/src/layout/AppShell.module.css`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 3.1: Add ProjectSwitcher component**

Create `web/src/components/project/ProjectSwitcher.tsx` with a compact native `<select>` and manage link:

```tsx
import { Link } from 'react-router-dom';

import { useProject } from '../../auth/ProjectContext';
import { useI18n } from '../../locale/useI18n';
import styles from './ProjectSwitcher.module.css';

export function ProjectSwitcher() {
  const { projects, currentProject, loading, error, selectProject } = useProject();
  const { t } = useI18n();

  return (
    <div className={styles.switcher} aria-label={t('project.switcher.label')}>
      <span className={styles.label}>{t('project.switcher.current')}</span>
      <select
        className={styles.select}
        value={currentProject?.code ?? ''}
        disabled={loading || projects.length === 0}
        onChange={(event) => selectProject(event.target.value)}
      >
        {projects.length === 0 ? (
          <option value="">{loading ? t('project.switcher.loading') : t('project.switcher.empty')}</option>
        ) : (
          projects.map((project) => (
            <option key={project.code} value={project.code}>
              {project.name} ({project.code})
            </option>
          ))
        )}
      </select>
      <Link className={styles.manage} to="/projects">
        {t('project.switcher.manage')}
      </Link>
      {error ? <span className={styles.error}>{error}</span> : null}
    </div>
  );
}
```

- [ ] **Step 3.2: Add ProjectSwitcher styles**

Create CSS matching existing restrained shell controls. Keep stable heights and no nested cards.

- [ ] **Step 3.3: Mount in AppShell**

Modify `web/src/layout/AppShell.tsx`:

```tsx
import { ProjectSwitcher } from '../components/project/ProjectSwitcher';
```

Place `<ProjectSwitcher />` before `<LanguageSwitcher />` inside `topbarActions`.

- [ ] **Step 3.4: Update AppShell CSS**

Ensure `.topbarActions` can hold switcher + language + session without overlap at mobile widths.

- [ ] **Step 3.5: Add locale keys**

Add `project.switcher.*` keys to both `en` and `zhCN`.

- [ ] **Step 3.6: Verify**

Run: `cd /home/lilin/trae-project/Optimus/web && npm run build`
Expected: PASS.

- [ ] **Step 3.7: Commit**

```bash
git add web/src/components/project web/src/layout/AppShell.tsx web/src/layout/AppShell.module.css web/src/locale/messages.ts
git commit -m "Add project switcher to app shell"
```

---

### Task 4: Project List Page

**Files:**
- Create: `web/src/view/project/ProjectsPage.tsx`
- Create: `web/src/view/project/ProjectsPage.module.css`
- Modify: `web/src/router/index.tsx`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 4.1: Build ProjectsPage**

Create a dense operational page with:
- list of accessible projects
- inline create form (`name`, `code`, `description`)
- per-row edit controls for `name` and `description`
- per-row delete button
- link to `/projects/:code/members`
- error banner text from API errors

Do not invent account/asset count columns; SP1 API does not return them.

- [ ] **Step 4.2: Add route**

Modify `web/src/router/index.tsx`:

```tsx
import { ProjectsPage } from '../view/project/ProjectsPage';
```

Add:

```tsx
{
  path: 'projects',
  element: <ProjectsPage />,
}
```

- [ ] **Step 4.3: Add locale keys**

Add `project.list.*` strings to both locale dictionaries.

- [ ] **Step 4.4: Verify**

Run: `cd /home/lilin/trae-project/Optimus/web && npm run build`
Expected: PASS.

- [ ] **Step 4.5: Commit**

```bash
git add web/src/view/project/ProjectsPage.tsx web/src/view/project/ProjectsPage.module.css web/src/router/index.tsx web/src/locale/messages.ts
git commit -m "Add project management page"
```

---

### Task 5: Project Members Page

**Files:**
- Create: `web/src/view/project/ProjectMembersPage.tsx`
- Create: `web/src/view/project/ProjectMembersPage.module.css`
- Modify: `web/src/router/index.tsx`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 5.1: Build ProjectMembersPage**

Create a page that:
- reads `code` from `useParams`
- calls `listProjectMembers(code)`
- supports adding a member by numeric `userId` and role select
- supports changing each member role
- supports removing a member
- links back to `/projects`

Keep the user selector as numeric input in SP3; searchable user picker is future polish.

- [ ] **Step 5.2: Add route**

Modify `web/src/router/index.tsx`:

```tsx
import { ProjectMembersPage } from '../view/project/ProjectMembersPage';
```

Add nested route:

```tsx
{
  path: 'projects/:code/members',
  element: <ProjectMembersPage />,
}
```

- [ ] **Step 5.3: Add locale keys**

Add `project.members.*` strings to both locale dictionaries.

- [ ] **Step 5.4: Verify**

Run: `cd /home/lilin/trae-project/Optimus/web && npm run build`
Expected: PASS.

- [ ] **Step 5.5: Commit**

```bash
git add web/src/view/project/ProjectMembersPage.tsx web/src/view/project/ProjectMembersPage.module.css web/src/router/index.tsx web/src/locale/messages.ts
git commit -m "Add project member management page"
```

---

### Task 6: SideNav Three-Section Refactor

**Files:**
- Modify: `web/src/components/nav/SideNav.tsx`
- Modify: `web/src/components/nav/SideNav.module.css`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 6.1: Refactor SideNav sections**

Implement:
- Overview: Dashboard, Projects if `hasMenu('project.menu')`
- Current Project: show selected project name/code from `useProject()`
- Current Project children: Cloud Accounts if `hasMenu('cloud.accounts.menu')`, Cloud Assets if `hasMenu('cloud.assets.menu')`
- System: current existing system children

For cloud links, use `/cloud/accounts` and `/cloud/assets`. These routes may be added by SP4; until then users without SP4 route will hit the existing router fallback behavior. If this is too rough during implementation, add disabled text rows instead and leave actual links to SP4.

- [ ] **Step 6.2: Update SideNav CSS**

Add section labels and disabled project state styles. Preserve current collapse behavior for System only unless the implementation is simpler with always-expanded sections.

- [ ] **Step 6.3: Add locale keys**

Add `shell.overview`, `shell.projects`, `shell.currentProject`, `shell.noProjectSelected`, `shell.cloudAccounts`, and `shell.cloudAssets`.

- [ ] **Step 6.4: Verify**

Run: `cd /home/lilin/trae-project/Optimus/web && npm run build`
Expected: PASS.

- [ ] **Step 6.5: Commit**

```bash
git add web/src/components/nav/SideNav.tsx web/src/components/nav/SideNav.module.css web/src/locale/messages.ts
git commit -m "Refactor navigation around project context"
```

---

### Task 7: Final Verification

**Files:**
- Verify only.

- [ ] **Step 7.1: Frontend build**

Run: `cd /home/lilin/trae-project/Optimus/web && npm run build`
Expected: PASS.

- [ ] **Step 7.2: Backend regression**

Run: `cd /home/lilin/trae-project/Optimus/server && GOCACHE=/tmp/go-build-cache go test ./...`
Expected: PASS.

- [ ] **Step 7.3: Manual smoke**

Run dev server:

```bash
cd /home/lilin/trae-project/Optimus/web && npm run dev -- --host 127.0.0.1
```

Smoke checks:
- login page still loads
- after login, topbar switcher renders
- `/projects` renders
- `/projects/:code/members` renders for an accessible project
- changing project updates localStorage key `optimus.currentProjectCode`

- [ ] **Step 7.4: Commit any final fixes**

Only commit if Step 7 found fixes.

---

## Out of Scope

- Cloud accounts/assets frontend pages and empty states (SP4).
- Searchable user picker for member add.
- Project restore UI.
- "All projects" aggregate mode.
- Project account/asset counters unless the backend API later exposes them.

## Implementation Notes

- Use existing table/card visual density from system pages; avoid landing-page or marketing layout.
- Do not add dependencies.
- Keep project context independent from auth context, but mount it only under authenticated routes.
- If `project.menu` is missing, hide `/projects` nav entry; route can still exist and rely on backend/RBAC for hard enforcement.
- SP2 backend requires `X-Project-Code` for creating cloud accounts; SP3 must set it globally but does not need cloud pages yet.
