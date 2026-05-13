# FinOps + DevOps Platform Roadmap Design

**Date:** 2026-05-12
**Status:** Approved

## Document Purpose

This document is a platform roadmap for `Optimus`, not an implementation spec.

Its purpose is to define:

- the intended platform direction
- the sequence of major capability phases
- the dependency relationship between those phases
- the decomposition rule for future design and planning work

It does not define implementation details such as data models, API contracts, storage layout, provider adapters, scheduling strategies, or package structure. Those decisions belong in later stage-specific specs.

It is also not the place to define frontend UX, billing integration behavior, or provider-specific implementation choices.

## Current Baseline

The current repository is still centered on the `system` domain and foundational platform governance capabilities. The active focus is on stabilizing:

- identity, authentication, authorization, and audit behavior
- module composition and service boundary cleanup
- the backend shell needed to support later platform expansion

At this stage, `Optimus` should not be described as already having a cloud control-plane core. Cloud account onboarding, unified asset management, DevOps execution, FinOps analysis, and automation remain future platform phases rather than established codebase capabilities.

## Platform Direction

`Optimus` is intended to evolve into a unified operations platform that connects governance, infrastructure visibility, execution, and cost management in one system.

The long-term direction is:

1. establish a reliable governance foundation
2. build a shared cloud resource foundation
3. expand that foundation into DevOps capabilities
4. expand that foundation into FinOps capabilities
5. close the loop with automation and policy-driven operations

The core product idea is that later operational and cost capabilities should not invent separate models of infrastructure state. They should build on a shared platform foundation and a common understanding of managed resources.

## Roadmap Phases

### Phase 0: System Foundation

Goal:
Establish the platform's baseline governance capabilities, including the system domain needed to manage users, roles, permissions, authentication, and auditability.

Why it comes first:
Every later platform capability depends on a stable identity, access, and traceability model.

Non-goal:
This phase is not where cloud management, deployment orchestration, or cost analysis begins.

### Phase 0.5: Audit and Governance Hardening

Goal:
Strengthen audit and governance capabilities so the platform can support higher-risk operational workflows with adequate traceability and control.

Why it comes next:
Before the platform starts managing cloud credentials, resources, or automated actions, it needs stronger governance maturity than a basic write-only audit trail.

Role in the roadmap:
This phase is a hardening step between foundational system capabilities and broader platform expansion.

### Phase 1: Cloud Resource Foundation

Goal:
Introduce the shared cloud resource foundation that future platform capabilities will depend on.

What this phase establishes:

- a way to connect the platform to external cloud environments
- a consistent internal view of managed cloud resources
- a common base for later operational and financial capabilities

Why it comes before DevOps and FinOps:
Both DevOps and FinOps need a stable object model for the infrastructure they operate on or analyze. Without that shared foundation, each future capability would drift toward its own incompatible representation of cloud resources.

### Phase 2: DevOps Enablement

Goal:
Build operational capabilities on top of the shared resource foundation so managed infrastructure can be acted on, not just observed.

Role in the roadmap:
This phase extends the platform from governance and visibility into controlled operational execution.

Dependency:
It depends on the prior existence of a shared resource foundation and governance controls.

Planning note:
After Phase 1 is stable, this phase can be planned independently from Phase 3. The roadmap lists them in sequence for communication clarity, not because one must be fully implemented before the other can be designed.

### Phase 3: FinOps Enablement

Goal:
Build financial governance and cost analysis capabilities on top of the same shared resource foundation.

Role in the roadmap:
This phase extends the platform from infrastructure visibility into cost visibility, allocation, and optimization.

Dependency:
It depends on the same shared foundation used by DevOps capabilities, so cost analysis and infrastructure operations remain aligned to the same resource model.

Planning note:
After Phase 1 is stable, this phase can also be planned independently from Phase 2 based on product priority and team capacity.

### Phase 4: Automation and Policy

Goal:
Introduce automation and policy-driven workflows that connect governance, infrastructure state, operational actions, and financial controls into a closed loop.

Role in the roadmap:
This phase turns the platform from a collection of managed capabilities into a system that can enforce intent and reduce manual operational overhead.

Dependency:
This phase should build on stable foundations from the earlier phases rather than define them itself.

## Decomposition Principle

This roadmap is intentionally high level.

When a phase becomes ready for execution, it must be broken into one or more dedicated design specs and implementation plans that reflect the codebase reality at that time.

The next artifact after this roadmap should always be a scoped phase design spec, not a repo-wide implementation plan.

Rules for decomposition:

- the roadmap does not serve as a build-ready design
- each implementation phase must have its own scoped spec
- each scoped spec must be validated against the then-current repository structure
- detailed technical choices must be made in those scoped specs, not backfilled into this roadmap

This rule prevents the roadmap from being misread as a frozen implementation decision record.

## Status Snapshot

Current focus:

- system-domain stabilization
- governance and audit hardening
- boundary cleanup in the backend architecture

Next likely platform focus:

- cloud resource foundation

Later platform directions:

- DevOps enablement
- FinOps enablement
- automation and policy
