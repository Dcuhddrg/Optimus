# Cloud Resource Foundation Design

> 2026-10-26 完整性说明：本文仍有仅标题的安全治理、同步生命周期、API 边界、交付切片及验证章节，不能视为完整实施契约。承接候选为 [账号资产设计](2026-05-13-cloud-account-asset-registry-design.md) 与 [项目隔离设计](2026-05-14-multi-project-cloud-isolation-design.md)，逐项映射尚待核实，跟进 [#3](https://github.com/Dcuhddrg/Optimus/issues/3)。原文与空标题保留，不补造决定。

**Date:** 2026-05-13

## Background

Phase 1 exists because the [roadmap](../../memory/platform-roadmap.md) separates shared cloud resource foundation work from the later DevOps, FinOps, and automation phases. The roadmap makes this sequence explicit: the platform must first establish a shared cloud resource base that later phases can depend on without redefining the core resource model.

This spec is the boundary-setting document for that first platform phase. Its job is to lock the scope before any deeper design work starts, so later phase specs can reuse the same foundation instead of re-litigating account, resource, and visibility semantics.

## Scope

This spec is limited to Phase 1 cloud account onboarding and shared resource visibility.

In scope:

- connecting the platform to external cloud accounts
- establishing a canonical internal view of managed cloud resources
- defining the minimum shared resource foundation that later phases can reuse

Out of scope:

- DevOps execution workflows
- FinOps analysis, attribution, budgeting, or optimization
- policy automation or closed-loop execution
- product UX beyond what is required to support the foundation boundary

## Non-goals

Phase 1 does not define:

- Phase 2 DevOps concerns
- Phase 3 FinOps concerns
- Phase 4 automation concerns
- provider-specific automation behavior beyond resource visibility and account onboarding
- full lifecycle management for every cloud resource type
- a backend-wide architecture rewrite unrelated to the cloud foundation

## Current Codebase Constraints

The current backend is still `system`-centered. The active repository layout is organized around `server/api/v1/system`, `server/router/system`, `server/service/system`, and `server/initialize`, not around a mature `cloud` domain.

At the time of this design, the architecture documentation described a `main -> initialize -> router -> api/handlers -> usecase/service -> infra/persistence` flow. This historical description is superseded by [ADR-002](../../memory/decisions/ADR-002-backend-structure-migration.md); current architecture requirements are in [backend layer rules](../../modules/backend-layer-rules.md).

The recent module-composition and service-boundary specs, [2026-05-11-system-module-composition-design.md](../archive/specs/2026-05-11-system-module-composition-design.md) and [2026-05-12-system-service-boundary-design.md](../archive/specs/2026-05-12-system-service-boundary-design.md), also show that the codebase is still tightening boundary discipline inside `system`. This Phase 1 spec must therefore anchor itself to the repository as it exists now, not to an imagined post-refactor cloud platform.

## Design Overview

Phase 1 establishes the shared cloud resource foundation that later platform work can build on without redefining what a cloud account is, how a managed resource is identified, or where the canonical view of that resource lives.

The design separates four related concerns that must not be collapsed into one opaque layer:

- **Cloud account connectivity**: the platform needs a way to attach to external cloud environments and keep those connections usable as a platform concern.
- **Resource visibility**: once connected, the platform needs a repeatable way to observe resources that exist in those environments.
- **Shared normalization**: discovered resources must be translated into one internal vocabulary that later phases can reuse.
- **Downstream consumption**: later DevOps and FinOps work should consume the shared foundation instead of re-describing account or resource identity.

The central design choice in Phase 1 is to treat the resource registry as the common foundation. Cloud accounts are the entry point, discovery is the observation mechanism, sync orchestration keeps the view current, and governance hooks consume the normalized result. That sequence matters because a connected account does not yet mean the platform has a usable internal model, and a discovered resource does not yet mean the platform has a stable shared record.

The minimum shared resource contract for later phases should be defined only at the level needed to keep the foundation reusable:

- a stable reference to the external cloud account or environment that owns the resource
- a provider-scoped resource identity that remains stable across syncs
- a normalized resource kind or category
- a scope marker such as region, zone, subscription, project, or equivalent cloud boundary
- a normalized metadata envelope for common descriptive fields
- a visibility or lifecycle status that tells consumers whether the resource is known, current, stale, or otherwise not yet reliable for downstream use

This contract is intentionally narrower than any final implementation model. Later phases may add richer attributes, but they must not redefine these shared anchors.

Phase 1 owns the following responsibilities:

- establish the external cloud connection boundary
- discover resources from connected environments
- normalize discovered resources into a shared internal registry
- coordinate refresh and reconciliation of that registry
- expose non-executing governance touchpoints that can read the shared foundation

Phase 1 intentionally defers:

- provider-specific execution behavior
- cost allocation and billing semantics
- policy automation and closed-loop remediation
- any final API contract or storage schema
- any package or directory layout decision that would overfit the design to the current code shape

## Domain Breakdown

### Cloud Account Management

What it does:

- establishes and maintains the platform relationship to an external cloud account or environment
- tracks whether a connection is usable for downstream discovery and sync
- serves as the entry point for the shared cloud foundation

What consumes it:

- resource discovery
- sync orchestration
- governance hooks that need account context
- later DevOps and FinOps phases that need a platform-level cloud attachment point

What it depends on:

- the existing `system` governance and identity baseline
- external cloud access credentials or equivalent trust material
- the shared notion of a cloud account anchor that the registry can reference

What it explicitly does not own:

- resource normalization
- the canonical registry of managed resources
- provider-specific discovery logic
- policy evaluation or remediation
- cost, billing, or execution workflows

### Cloud Resource Discovery

What it does:

- inspects connected cloud environments for managed resources
- collects the information needed to populate the shared internal view
- translates raw provider visibility into data that can be normalized

What consumes it:

- the unified resource registry
- sync orchestration
- governance hooks that need current resource visibility
- later Phase 2 and Phase 3 planning work that depends on a stable inventory surface

What it depends on:

- cloud account management for access to connected environments
- provider read access and cloud-specific observation capabilities
- the normalization rules that define what the registry expects

What it explicitly does not own:

- the durable registry itself
- connection lifecycle management
- downstream operational actions on resources
- cost attribution or policy automation
- final provider adapter shapes or implementation packaging

### Unified Resource Registry

What it does:

- holds the canonical internal view of managed cloud resources
- normalizes identity, scope, and descriptive metadata into one shared representation
- acts as the shared foundation for later DevOps and FinOps work

What consumes it:

- governance hooks
- sync orchestration feedback paths
- later DevOps and FinOps specs
- cross-cutting reporting or search surfaces that need one resource vocabulary

What it depends on:

- cloud account management for account anchors
- cloud resource discovery for observed resource facts
- sync orchestration for keeping records current and reconciling changes

What it explicitly does not own:

- external cloud connectivity
- discovery execution
- policy engines
- billing logic
- task runners or automated remediation

### Sync Orchestration

What it does:

- coordinates initial and repeated refresh of cloud account and resource data
- sequences the movement from discovered resource facts into the registry
- defines when the shared view is considered fresh enough for consumers

What consumes it:

- cloud account management workflows
- the unified resource registry
- governance hooks that care about freshness or staleness
- later phase planning that needs a repeatable refresh model

What it depends on:

- cloud account management
- cloud resource discovery
- the unified resource registry write path
- the platform's existing operational control surface

What it explicitly does not own:

- the discovery algorithm itself
- the registry contract itself
- provider-specific remediation
- policy-driven automation
- any future workflow runner abstraction

### Governance Hooks

What it does:

- provides read-side integration points for governance-aware consumers
- allows the shared foundation to be validated, annotated, or inspected without turning Phase 1 into an automation phase
- keeps governance concerns close to the shared registry while staying non-executing

What consumes it:

- current governance-oriented platform logic
- future DevOps and FinOps specs that need guardrail-aware read access
- review, audit, and visibility surfaces that depend on the shared foundation

What it depends on:

- the unified resource registry
- cloud account context
- the platform's existing governance baseline

What it explicitly does not own:

- policy engine execution
- remediation or enforcement workflows
- automation scheduling
- cost allocation decisions
- provider-specific control operations

## Phase 1 Stability and Handoff Criteria

Phase 1 is not complete unless it defines all three roadmap-level outcomes:

- a platform-level way to connect to external cloud environments
- a consistent internal view of managed resources
- a shared resource foundation that later DevOps and FinOps specs can use without redefining core resource identity

Phase 1 is stable enough for later Phase 2 or Phase 3 specs when those downstream specs can reuse the shared foundation without redefining account and resource semantics or reopening the Phase 1 resource boundary.

- the shared cloud resource vocabulary is fixed enough that later specs can reuse it without redefining account and resource semantics
- cloud account onboarding and shared resource visibility use the same account/resource vocabulary and boundary assumptions that later Phase 2 and Phase 3 specs can reuse without redefining the shared foundation
- later phase specs can assume the shared foundation instead of reopening the Phase 1 resource boundary
- any changes to the shared resource foundation would be treated as a foundation revision, not as a Phase 2 or Phase 3 local tweak

At that point, Phase 2 and Phase 3 can start from the shared foundation instead of spending their first pass reconstructing the underlying cloud resource model.

## Security and Governance Constraints

## Sync and Lifecycle Model

## API and Module Boundaries

## Delivery Slices

## Testing and Verification Strategy
