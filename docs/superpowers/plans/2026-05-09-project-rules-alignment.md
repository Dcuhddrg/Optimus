# Project Rules Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update `.codex/project_rules.md` so it remains a shared Claude Code / Codex contract while aligning workflow, verification, and commit guidance with the repository’s current execution reality.

**Architecture:** This is a documentation-only change with one target file. The implementation keeps the existing section skeleton intact and performs in-place edits: preserve the useful boundary and output guidance, update misleading workflow and verification defaults, and replace the outdated commit-format rule with the repository’s current Lore-oriented guidance.

**Tech Stack:** Markdown documentation, repository-local rules in `.codex/`, existing docs under `docs/superpowers/`.

---

## Scope Notes

- This plan modifies `.codex/project_rules.md` only.
- Do not rewrite the file into a new top-level structure.
- Do not modify `docs/architecture/current-architecture.md`, `AGENTS`-style files, implementation code, or CI/config files.
- Keep the document as a shared contract for Claude Code and Codex; do not make it agent-specific.
- Preserve the requirement for Chinese-oriented communication/output guidance.

## File Map

### Modify

- `.codex/project_rules.md` — shared repository rules contract edit path; in this repo it is a symlink to `../.claude/project_rules.md`
- `.claude/project_rules.md` — tracked content path that actually records the in-place rules update in git

### Do Not Modify

- `docs/architecture/current-architecture.md` — remains the architecture reference only
- `docs/superpowers/specs/2026-05-09-project-rules-alignment-design.md` — approved design reference only
- `server/**`, `web/**`, `deploy/**` — no code or runtime changes in this task

## Task 1: Align Workflow Guidance Without Changing the Document Skeleton

**Files:**
- Modify: `.codex/project_rules.md`
- Modify: `.claude/project_rules.md`

- [ ] **Step 1: Review the current workflow and execution sections**

Run:

```bash
sed -n '1,240p' .codex/project_rules.md
```

Expected: the file still contains the current `BOUNDARIES`, `DEFAULT EXECUTION`, and `WORKFLOW` sections, including the outdated fixed chain `brainstorming → writing-plans → executing-plans`.

- [ ] **Step 2: Rewrite the workflow section in place**

Edit `.codex/project_rules.md` so `WORKFLOW` says, in substance:

- non-trivial tasks should still start with `brainstorming`
- approved specs should still flow into `writing-plans`
- implementation after planning may use:
  - `subagent-driven-development`
  - `executing-plans`

Do not rename the section, and do not add a new top-level “modes” or “agent” chapter.

- [ ] **Step 3: Soften overlapping execution language**

Edit `DEFAULT EXECUTION` so it remains useful at the repository level but stops competing with upstream agent-specific dialogue rules.

Keep these ideas:

- small, incremental edits
- verify as you go
- fit existing patterns
- do not expand confirmed front-end scope

Weaken or remove purely conversational mandates if they are redundant with higher-level agent rules.

- [ ] **Step 4: Re-check section shape and workflow content**

Run:

```bash
grep -nE '^## (BOUNDARIES|DEFAULT EXECUTION|WORKFLOW|VERIFICATION|OUTPUT CONTRACT|ARCHITECTURE REFERENCE)$' .codex/project_rules.md
grep -n 'brainstorming' .codex/project_rules.md
grep -n 'subagent-driven-development' .codex/project_rules.md
grep -n 'executing-plans' .codex/project_rules.md
```

Expected:

- all original top-level sections still exist
- `brainstorming` remains present
- `subagent-driven-development` and `executing-plans` are both present
- the file no longer implies `executing-plans` is the only default implementation endpoint

- [ ] **Step 5: Commit**

```bash
git add .codex/project_rules.md
git add .claude/project_rules.md
git commit -m "使共享工作流规则与当前执行路径保持一致" \
  -m "Constraint: Must preserve the current document skeleton and shared-agent posture" \
  -m "Rejected: Rewrite project rules into a new governance structure | too much scope for a consistency update" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Keep repository workflow guidance descriptive enough to be useful without hardcoding a single implementation skill as the only endpoint" \
  -m "Tested: sed -n '1,240p' .codex/project_rules.md" \
  -m "Tested: grep -nE '^## (BOUNDARIES|DEFAULT EXECUTION|WORKFLOW|VERIFICATION|OUTPUT CONTRACT|ARCHITECTURE REFERENCE)$' .codex/project_rules.md"
```

## Task 2: Tighten Verification Guidance to Match Current Front-End Practice

**Files:**
- Modify: `.codex/project_rules.md`
- Modify: `.claude/project_rules.md`

- [ ] **Step 1: Identify the current verification baseline text**

Run:

```bash
grep -nE 'npm install|npm run build|npm run dev|go test ./\.\.\.|go build ./\.\.\.' .codex/project_rules.md
```

Expected: the file still lists `cd web && npm install` as part of the front-end default verification baseline.

- [ ] **Step 2: Update the front-end verification baseline**

Edit `.codex/project_rules.md` so the default front-end verification baseline becomes:

- `cd web && npm run build`

Move `npm install` out of the default baseline and restate it as conditional environment preparation only when:

- dependencies are missing
- lockfile or dependency state changed
- the local environment is incomplete

Keep `npm run dev` and “dev server reachable” as additional checks for tasks involving runtime behavior.

- [ ] **Step 3: Preserve non-front-end verification rules**

Ensure the backend baseline and documentation/rules verification guidance remain intact:

- `cd server && go test ./...`
- `cd server && go build ./...`
- structure/reference/boundary checks for pure docs/rules work

Do not broaden verification scope beyond what the spec calls for.

- [ ] **Step 4: Re-check verification wording**

Run:

```bash
grep -nE 'npm install|npm run build|npm run dev|go test ./\.\.\.|go build ./\.\.\.' .codex/project_rules.md
sed -n '/前端代码改动默认验证基线：/,/如果任务涉及前端运行效果，再追加：/p' .codex/project_rules.md
```

Expected:

- `npm run build` remains in the front-end baseline
- `npm install` no longer appears inside the default front-end baseline list; if it remains, it appears only as conditional environment-preparation guidance
- backend and runtime verification references still exist

- [ ] **Step 5: Commit**

```bash
git add .codex/project_rules.md
git add .claude/project_rules.md
git commit -m "使前端验证默认值贴合当前实际做法" \
  -m "Constraint: Verification guidance must stay practical for an already-initialized workspace" \
  -m "Rejected: Keep npm install in the default verification baseline | it conflates environment setup with validating the actual change" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Treat dependency installation as conditional environment prep, not the default proof that a front-end change is correct" \
  -m "Tested: grep -nE 'npm install|npm run build|npm run dev|go test ./\\.\\.\\.|go build ./\\.\\.\\.' .codex/project_rules.md"
```

## Task 3: Replace the Outdated Commit Rule While Keeping Chinese Description Guidance

**Files:**
- Modify: `.codex/project_rules.md`
- Modify: `.claude/project_rules.md`

- [ ] **Step 1: Identify the current commit-format block**

Run:

```bash
grep -nE 'Git Commit|feat\\(web\\)|中文描述|Lore' .codex/project_rules.md
```

Expected: the file still contains the old `feat(web): 中文描述`-style guidance and does not yet express the current Lore-oriented structure correctly.

- [ ] **Step 2: Update the commit guidance in place**

Edit `.codex/project_rules.md` so the commit rules say, in substance:

- commit intent lines should remain Chinese descriptions
- commit messages must follow the repository’s current Lore protocol structure
- old Conventional Commit examples such as `feat(web): 中文描述` are removed as the default rule

If an example is kept, use a Lore-style shape with a Chinese intent line and trailers rather than the old `type(scope): ...` format.

- [ ] **Step 3: Preserve the rest of the output contract**

Keep the surrounding `OUTPUT CONTRACT` behavior intact:

- Chinese conversation/output
- concise, scannable reporting
- conclusion before process

This task updates commit guidance only; it should not rewrite the whole section.

- [ ] **Step 4: Re-check the output contract block**

Run:

```bash
grep -nE 'feat\\(web\\)|Lore|中文描述|Output Contract|OUTPUT CONTRACT' .codex/project_rules.md
```

Expected:

- the old `feat(web)` example is gone
- Chinese-description guidance remains
- Lore-oriented guidance is present

- [ ] **Step 5: Commit**

```bash
git add .codex/project_rules.md
git add .claude/project_rules.md
git commit -m "使提交规则与当前仓库风格保持一致" \
  -m "Constraint: Chinese intent descriptions must remain part of the shared contract" \
  -m "Rejected: Keep the old feat(scope) example as the default | it conflicts with the repository's current Lore protocol practice" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Prefer one clear commit convention in project rules; mixed examples quickly become misleading" \
  -m "Tested: grep -nE 'feat\\(web\\)|Lore|中文描述|Output Contract|OUTPUT CONTRACT' .codex/project_rules.md"
```

## Task 4: Final Consistency Review of `.codex/project_rules.md`

**Files:**
- Modify: `.codex/project_rules.md`
- Modify: `.claude/project_rules.md`

- [ ] **Step 1: Read the final file top to bottom**

Run:

```bash
sed -n '1,260p' .codex/project_rules.md
```

Expected: the file still reads as one shared repository contract with the original section skeleton, updated workflow guidance, corrected verification defaults, and current commit rules.

- [ ] **Step 2: Verify the scope stayed narrow**

Run:

```bash
git status --short -- .codex/project_rules.md .claude/project_rules.md
```

Expected: the scoped output shows only intentionally changed tracked paths. In practice, `.claude/project_rules.md` is the required dirty tracked content path; `.codex/project_rules.md` may remain absent because it is only the symlinked edit/read alias.

- [ ] **Step 3: Verify no whitespace or formatting regressions**

Run:

```bash
git diff --check -- .claude/project_rules.md
```

Expected: no whitespace errors in the tracked rules content path.

- [ ] **Step 4: Record the residual risk for final handoff**

Include this in the eventual implementation handoff if still true:

- the file is aligned to current practice now, but future agent/workflow changes can still make it drift again if not maintained alongside process changes

- [ ] **Step 5: Commit**

If any final wording cleanup was needed after review, commit it with a final Lore-style message describing the consistency cleanup. If no cleanup edits were needed, do not create an extra commit.
