# Codex Project Rules Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restructure Codex-facing guidance so `.codex/project_rules.md` becomes a pure behavior contract and all repository architecture facts move into `docs/architecture/current-architecture.md`.

**Architecture:** The change is intentionally documentation-only. The implementation keeps one file for agent behavior rules and one file for current repository facts, linked by a short reference but with no duplicated body content. Verification focuses on boundary clarity, content completeness, and zero code churn outside the two target files.

**Tech Stack:** Markdown documentation, existing Go backend repository context, repository-local Codex rules under `.codex/`

---

## File Map

### Create

- `docs/architecture/current-architecture.md` — authoritative description of current repository scope, layout, backend architecture, runtime flow, tech stack, public endpoints, verification commands, and non-goals

### Modify

- `.codex/project_rules.md` — Codex behavior contract only, with architecture details removed and replaced by a short reference

### Do Not Modify

- `server/**` — backend code and tests
- `README.md` — existing quick-start note remains unchanged for this task
- `docs/superpowers/**` except this plan/spec pair — historical design and plan documents remain unchanged

## Task 1: Rewrite `.codex/project_rules.md` as a pure Codex behavior contract

**Files:**
- Modify: `.codex/project_rules.md`

- [ ] **Step 1: Write the failing content checklist**

Create a temporary checklist in your notes with these expected conditions:

```text
- file no longer contains repository layout tree
- file no longer contains detailed tech stack table
- file contains Boundaries section
- file contains Default Execution section
- file contains Verification section
- file contains Output Contract section
- file contains Architecture Reference section
```

- [ ] **Step 2: Verify the current file fails the checklist**

Run: `sed -n '1,260p' .codex/project_rules.md`
Expected: FAIL the checklist because the file still contains detailed architecture and tech stack content.

- [ ] **Step 3: Write the minimal replacement content**

Replace the file with a Codex-only rules document that includes:

- a short title and purpose statement
- `Boundaries`
- `Default Execution`
- `Verification`
- `Output Contract`
- `Architecture Reference`

The replacement should explicitly:

- strengthen “do not go beyond requested scope”
- strengthen “default to executing directly when the task is clear”
- keep Chinese conversation/output expectations
- reference `docs/architecture/current-architecture.md` instead of restating architecture facts

- [ ] **Step 4: Verify the rewritten file satisfies the checklist**

Run: `sed -n '1,260p' .codex/project_rules.md`
Expected: PASS the checklist with no embedded architecture tree or full tech stack table.

- [ ] **Step 5: Commit**

```bash
git add .codex/project_rules.md
git commit -m "docs(codex): tighten project rules contract"
```

## Task 2: Create the current architecture reference document

**Files:**
- Create: `docs/architecture/current-architecture.md`

- [ ] **Step 1: Write the failing content checklist**

Create a temporary checklist in your notes with these expected sections:

```text
- Scope
- Repository Layout
- Backend Architecture
- Runtime Flow
- Tech Stack
- Public Endpoints & Verification
- Non-goals
```

- [ ] **Step 2: Verify the document does not exist yet**

Run: `test -f docs/architecture/current-architecture.md && echo exists || echo missing`
Expected: `missing`

- [ ] **Step 3: Write the architecture document**

Create `docs/architecture/current-architecture.md` with these sections:

- `Scope`
  - `server/` is the current main development surface
  - `web/` remains a placeholder
  - current stage is a runnable Go backend skeleton, not a full business system
- `Repository Layout`
  - key directories and one-line responsibilities only
- `Backend Architecture`
  - responsibilities of `main`, `core`, `initialize`, `router`, `api`, `service`, `model/common`, `utils`, `global`
- `Runtime Flow`
  - config file path
  - `OPTIMUS_` environment override pattern
  - optional Redis/Casbin initialization
  - router build and HTTP startup flow
- `Tech Stack`
  - Go 1.23, Gin, GORM, MySQL, Viper, Zap, Redis, Casbin, JWT, testify
- `Public Endpoints & Verification`
  - `/api/v1/health`
  - `/api/v1/ping`
  - `/api/v1/base/ping`
  - `cd server && go test ./...`
  - `cd server && go build ./...`
  - `cd server && go run .`
  - temporary port override command
- `Non-goals`
  - no full auth system
  - no user/role/menu domain
  - no Swagger
  - no code generator

- [ ] **Step 4: Verify the architecture document contains all required sections**

Run: `grep -nE '^## (Scope|Repository Layout|Backend Architecture|Runtime Flow|Tech Stack|Public Endpoints & Verification|Non-goals)$' docs/architecture/current-architecture.md`
Expected: seven matching section headers.

- [ ] **Step 5: Commit**

```bash
git add docs/architecture/current-architecture.md
git commit -m "docs(architecture): add current repository overview"
```

## Task 3: Verify separation of responsibilities between the two documents

**Files:**
- Modify: `.codex/project_rules.md`
- Modify: `docs/architecture/current-architecture.md`

- [ ] **Step 1: Check for duplicated architecture-heavy content in the rules file**

Run: `grep -nE 'Repository Layout|Tech Stack|GET /api/v1/health|server/config|server/core' .codex/project_rules.md`
Expected: only a short architecture reference line, or no matches for detailed architecture markers.

- [ ] **Step 2: Check that the architecture document does not include behavior-rule sections**

Run: `grep -nE '^## (Boundaries|Default Execution|Verification|Output Contract|Architecture Reference)$' docs/architecture/current-architecture.md || true`
Expected: no matches.

- [ ] **Step 3: Make any minimal wording adjustments needed**

If either check fails, adjust wording so:

- `.codex/project_rules.md` remains behavior-only
- `docs/architecture/current-architecture.md` remains repository-facts-only

- [ ] **Step 4: Re-run both separation checks**

Run:

```bash
grep -nE 'Repository Layout|Tech Stack|GET /api/v1/health|server/config|server/core' .codex/project_rules.md
grep -nE '^## (Boundaries|Default Execution|Verification|Output Contract|Architecture Reference)$' docs/architecture/current-architecture.md || true
```

Expected:

- first command shows no detailed architecture dump in the rules file
- second command shows no behavior-rule sections in the architecture doc

- [ ] **Step 5: Commit**

```bash
git add .codex/project_rules.md docs/architecture/current-architecture.md
git commit -m "docs(codex): separate rules from architecture"
```

## Task 4: Final review and handoff

**Files:**
- Modify: `.codex/project_rules.md`
- Modify: `docs/architecture/current-architecture.md`

- [ ] **Step 1: Review the final rules document**

Run: `sed -n '1,260p' .codex/project_rules.md`
Expected: concise Codex behavior contract with stronger constraints and stronger default execution guidance.

- [ ] **Step 2: Review the final architecture document**

Run: `sed -n '1,260p' docs/architecture/current-architecture.md`
Expected: current-state repository and backend architecture reference with no workflow-rule content.

- [ ] **Step 3: Verify no unintended project files were changed**

Run: `git status --short`
Expected: only `.codex/project_rules.md`, `docs/architecture/current-architecture.md`, and the current spec/plan documents are changed.

- [ ] **Step 4: Summarize remaining risks**

Document these risks in the final handoff if still true:

- future drift if architecture facts are copied back into `.codex/project_rules.md`
- current architecture doc may need updating whenever backend structure materially changes

- [ ] **Step 5: Commit**

```bash
git add .codex/project_rules.md docs/architecture/current-architecture.md
git commit -m "docs(codex): split rules from architecture facts"
```
