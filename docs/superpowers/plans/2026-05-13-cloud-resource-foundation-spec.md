> 历史路径说明：本计划中 `docs/architecture/current-architecture.md` 和 `docs/roadmaps/finops-devops-platform-roadmap.md` 为当时路径，已迁移；现行架构见 [系统架构](../../relations/repo-profile.md) 、[后端分层规则](../../modules/backend-layer-rules.md) 和 [平台路线图](../../memory/long-term/platform-roadmap.md)。历史命令保留原样，不应作为当前操作步骤。

# Cloud Resource Foundation Spec Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a scoped `Phase 1` design spec for the cloud resource foundation that is aligned with the project roadmap and the current Optimus codebase.

**Architecture:** This plan is docs-first by design. It does not implement `Phase 1` directly. Instead, it produces the next required artifact after the roadmap: a dedicated design spec that translates the roadmap's high-level direction into a phase-scoped backend design aligned with the existing `system`-centered architecture. The spec must stay focused on `Phase 1` and avoid leaking into `Phase 2`, `Phase 3`, or `Phase 4` implementation detail.

**Tech Stack:** Markdown documentation, existing project docs under `docs/`, Go backend structure under `server/`, codebase discovery via `find`/`sed`/MCP search, and the brainstorming + spec review workflow.

---

## File Map

| File | Change |
|------|--------|
| `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md` | Create the scoped `Phase 1` design spec |
| `docs/roadmaps/finops-devops-platform-roadmap.md` | Reference only; do not modify unless the new spec reveals a roadmap contradiction |
| `docs/architecture/current-architecture.md` | Reference current repository baseline and extension constraints |
| `docs/superpowers/specs/2026-05-11-system-module-composition-design.md` | Reference existing backend composition direction |
| `docs/superpowers/specs/2026-05-12-system-service-boundary-design.md` | Reference current service-boundary direction |
| `server/` subtree | Read-only reference to verify current backend layout and naming patterns |

---

### Task 1: Establish the spec boundary and source constraints

**Files:**
- Create: `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`
- Reference: `docs/roadmaps/finops-devops-platform-roadmap.md`
- Reference: `docs/architecture/current-architecture.md`
- Reference: `docs/superpowers/specs/2026-05-11-system-module-composition-design.md`
- Reference: `docs/superpowers/specs/2026-05-12-system-service-boundary-design.md`

- [ ] **Step 1: Confirm the source documents**

Run:

```bash
sed -n '1,220p' docs/roadmaps/finops-devops-platform-roadmap.md
sed -n '1,220p' docs/architecture/current-architecture.md
sed -n '1,220p' docs/superpowers/specs/2026-05-11-system-module-composition-design.md
sed -n '1,220p' docs/superpowers/specs/2026-05-12-system-service-boundary-design.md
```

Expected:
- The roadmap says the next artifact after the roadmap is a phase-scoped design spec
- The current architecture is still `system`-centered
- Existing backend direction already emphasizes composition and boundary discipline

- [ ] **Step 2: Map the current backend layout**

Run:

```bash
find server -maxdepth 3 -type d | sort
```

Expected:
- No mature `cloud` domain currently exists
- The current layout is sufficient to anchor the spec to real repository constraints

- [ ] **Step 3: Write the spec skeleton**

Add this initial structure to `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`:

```markdown
# Cloud Resource Foundation Design

**Date:** 2026-05-13

## Background

## Scope

## Non-goals

## Current Codebase Constraints

## Design Overview

## Domain Breakdown

## Phase 1 Stability and Handoff Criteria

## Security and Governance Constraints

## Sync and Lifecycle Model

## API and Module Boundaries

## Delivery Slices

## Testing and Verification Strategy
```

- [ ] **Step 4: Fill the boundary-setting sections first**

Write the `Background`, `Scope`, `Non-goals`, and `Current Codebase Constraints` sections before writing any design detail.

Minimum content requirements:

- `Background` explains why `Phase 1` exists in the roadmap
- `Scope` limits the spec to cloud account onboarding and shared resource visibility
- `Non-goals` explicitly excludes `Phase 2`, `Phase 3`, and `Phase 4` concerns
- `Current Codebase Constraints` ties the design back to the actual `system`-centric backend and ongoing boundary cleanup
- `Phase 1 Stability and Handoff Criteria` defines what must be true before later `Phase 2` or `Phase 3` specs can start without redefining the shared resource foundation

- [ ] **Step 5: Verify the spec starts with the correct boundary**

Run:

```bash
sed -n '1,120p' docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md
```

Expected:
- The top of the document is complete and clearly phase-scoped
- No implementation details appear before scope and constraints are set

---

### Task 2: Define the Phase 1 domain model at design level

**Files:**
- Modify: `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`
- Reference: `docs/roadmaps/finops-devops-platform-roadmap.md`
- Reference: `server/`

- [ ] **Step 1: Split `Phase 1` into bounded design units**

The `Domain Breakdown` section should define at least these units:

```markdown
- Cloud Account Management
- Cloud Resource Discovery
- Unified Resource Registry
- Sync Orchestration
- Governance Hooks
```

For each unit, answer:

- what it does
- what consumes it
- what it depends on
- what it explicitly does not own

- [ ] **Step 2: Keep the design at roadmap-following, not code-finalizing, depth**

Write the design so it is specific enough for later planning, but do not lock in:

- concrete table schemas
- exact API request and response shapes
- provider-specific adapters
- final package or directory names

- [ ] **Step 3: Write the `Design Overview` and `Domain Breakdown` sections**

These sections should explain:

- how cloud accounts and resource visibility relate
- why a shared resource registry is the common foundation
- the minimum shared resource identity and normalization contract that later `Phase 2` and `Phase 3` work can depend on
- which responsibilities belong inside `Phase 1`
- which concerns are intentionally deferred

- [ ] **Step 4: Add a Phase 1 acceptance check for the shared foundation**

The spec must explicitly state that `Phase 1` is not complete unless it defines all three roadmap-level outcomes:

- a platform-level way to connect to external cloud environments
- a consistent internal view of managed resources
- a shared resource foundation that later DevOps and FinOps specs can use without redefining core resource identity

- [ ] **Step 5: Verify no `Phase 2+` scope leaked in**

Run:

```bash
grep -nE 'pipeline|runner|cost allocation|billing|policy engine|automation workflow' docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md
```

Expected:
- Either no matches, or only explicit non-goal references

---

### Task 3: Define governance, lifecycle, and integration constraints

**Files:**
- Modify: `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`
- Reference: `docs/roadmaps/finops-devops-platform-roadmap.md`
- Reference: `docs/architecture/current-architecture.md`

- [ ] **Step 1: Write the `Security and Governance Constraints` section**

Cover at minimum:

- credentials are sensitive platform inputs
- all cloud access must remain compatible with existing governance expectations
- auditability must be preserved for high-risk actions
- the spec must not assume weaker controls than the current governance direction
- the spec must define who can onboard cloud accounts, who can view discovered resources, and at what scope those permissions apply

- [ ] **Step 2: Write the `Sync and Lifecycle Model` section**

Describe at design level:

- how resources move from external cloud state into platform visibility
- what “current visibility” versus “historical reference” means conceptually
- what kinds of lifecycle transitions later implementation must handle

Do not write:

- cron expressions
- batch sizes
- exact persistence fields
- exact retry algorithms

- [ ] **Step 3: Write the `API and Module Boundaries` section**

Describe:

- which user-facing capability areas will exist in `Phase 1`
- which backend seams the implementation will likely require
- where the design must align with existing backend composition rules
- how access scope and authorization boundaries apply to cloud accounts and shared resource visibility

Keep this at interface-boundary level. Do not assign final code symbols or exact handler/store method names.

- [ ] **Step 4: Verify the document still reads as design, not implementation**

Run:

```bash
sed -n '1,260p' docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md
```

Expected:
- The document describes boundaries, responsibilities, and lifecycle concepts
- It does not read like a migration, API, or code patch checklist

---

### Task 4: Define downstream delivery slices and verification expectations

**Files:**
- Modify: `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`

- [ ] **Step 1: Write the `Delivery Slices` section**

Break `Phase 1` into implementation-oriented slices that can later become separate plans. Use a structure like:

```markdown
1. Cloud account onboarding slice
2. Resource discovery slice
3. Unified resource registry slice
4. Governance and audit integration slice
```

Each slice should say:

- what user-visible outcome it enables
- what it depends on
- what it deliberately leaves for a later slice

- [ ] **Step 2: Write the `Phase 1 Stability and Handoff Criteria` section**

This section must translate the roadmap's `Phase 1` stability gate into spec-level completion language.

It should answer:

- what “stable enough to start Phase 2 or Phase 3 specs” means for `Phase 1`
- which capabilities must already be clear or present before later phases can plan against them
- which unresolved items are still acceptable without blocking later phase design

It must not answer this with implementation checklists or production rollout criteria.

- [ ] **Step 3: Write the `Testing and Verification Strategy` section**

Cover at least:

- unit and boundary tests expected from later implementation
- integration points that will need explicit verification
- migration or rollout risks that future plans must account for

Keep this at test-shape level. Do not enumerate concrete test cases unless they are necessary to avoid ambiguity.

- [ ] **Step 4: Verify the spec is planning-ready**

Run:

```bash
grep -nE 'TODO|TBD|placeholder' docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md
```

Expected:
- No matches

---

### Task 5: Review, user approval, and handoff to the real implementation plan

**Files:**
- Modify: `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`
- Reference: `docs/superpowers/plans/2026-05-13-cloud-resource-foundation-spec.md`

- [ ] **Step 1: Run the spec review loop**

Dispatch a reviewer subagent using the brainstorming skill's `spec-document-reviewer` prompt at `/home/lilin/.codex/superpowers/skills/brainstorming/spec-document-reviewer-prompt.md` with:

- spec path: `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`
- roadmap path: `docs/roadmaps/finops-devops-platform-roadmap.md`

Expected:
- Reviewer either approves the spec or returns actionable issues

- [ ] **Step 2: Fix review issues and re-run if needed**

Repeat until:

- the reviewer returns `Approved`, or
- three review loops have failed and a human decision is needed

- [ ] **Step 3: Commit the finished spec**

Run:

```bash
git add docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md
git commit -m "Define the Phase 1 cloud resource foundation scope" \
  -m "Document the scoped design for cloud resource onboarding and shared resource visibility so the roadmap can hand off into a concrete phase design artifact." \
  -m "Constraint: Must align with the current system-centered backend and the roadmap's phase boundaries" \
  -m "Rejected: Combine Phase 1 with later DevOps or FinOps design work | would collapse multiple roadmap phases into one spec" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Keep this spec limited to Phase 1 and defer later platform layers to separate specs" \
  -m "Tested: Spec-document review loop approved the final draft" \
  -m "Not-tested: No runtime verification required for this documentation change"
```

Expected:
- The spec is committed as a standalone design artifact

- [ ] **Step 4: Ask the user to review the spec**

Use this exact handoff:

```text
Spec written and committed to `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`. Please review it and let me know if you want to make any changes before we start writing out the implementation plan.
```

- [ ] **Step 5: If the user requests changes, revise and re-approve the spec**

If the user asks for changes:

- update `docs/superpowers/specs/2026-05-13-cloud-resource-foundation-design.md`
- rerun the spec review loop
- create a follow-up commit that records the requested adjustment
- ask for user review again using the same review gate message

Only continue once the user explicitly approves the updated spec.

- [ ] **Step 6: After user approval, create the actual Phase 1 implementation plan**

Only after the user approves the `Phase 1` spec:

- invoke `superpowers:writing-plans` again
- write a separate implementation plan for the approved `Phase 1` spec
- do not reuse this spec-authoring plan as the implementation plan
