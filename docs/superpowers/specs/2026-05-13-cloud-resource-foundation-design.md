# Cloud Resource Foundation Design

**Date:** 2026-05-13

## Background

Phase 1 exists because the [roadmap](../../roadmaps/finops-devops-platform-roadmap.md) separates shared cloud resource foundation work from the later DevOps, FinOps, and automation phases. The roadmap makes this sequence explicit: the platform must first establish a shared cloud resource base that later phases can depend on without redefining the core resource model.

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

The current architecture documentation describes a `main -> initialize -> router -> api/handlers -> usecase/service -> infra/persistence` flow in [current-architecture.md](../../architecture/current-architecture.md), which means any new cloud foundation work must fit the existing system-oriented composition path rather than assume a new top-level runtime model already exists.

The recent module-composition and service-boundary specs, [2026-05-11-system-module-composition-design.md](./2026-05-11-system-module-composition-design.md) and [2026-05-12-system-service-boundary-design.md](./2026-05-12-system-service-boundary-design.md), also show that the codebase is still tightening boundary discipline inside `system`. This Phase 1 spec must therefore anchor itself to the repository as it exists now, not to an imagined post-refactor cloud platform.

## Design Overview

## Domain Breakdown

## Phase 1 Stability and Handoff Criteria

Phase 1 is stable enough for later Phase 2 or Phase 3 specs when those downstream specs can reuse the shared foundation without redefining account and resource semantics or reopening the Phase 1 resource boundary.

- the shared cloud resource vocabulary is fixed enough that later specs can reuse it without redefining account and resource semantics
- cloud account onboarding and shared resource visibility are represented by the same foundation model across the repo
- later phase specs can assume the shared foundation instead of reopening the Phase 1 resource boundary
- any changes to the shared resource foundation would be treated as a foundation revision, not as a Phase 2 or Phase 3 local tweak

At that point, Phase 2 and Phase 3 can start from the shared foundation instead of spending their first pass reconstructing the underlying cloud resource model.

## Security and Governance Constraints

## Sync and Lifecycle Model

## API and Module Boundaries

## Delivery Slices

## Testing and Verification Strategy
