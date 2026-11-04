# GVA Response Alignment Implementation Plan

> **For agentic workers:** Execute one module at a time; verify each real route and its React caller before moving to the next module.

**Goal:** Replace Optimus's mixed JSON response protocols with the GVA `{code,data,msg}` contract while preserving existing business routes.

**Architecture:** The common response package owns GVA response helpers. Each module's API emits the same envelope; the React API client unwraps it and distinguishes business failures from network and parsing failures. CSV success and health/ping remain explicit content-type exceptions.

**Tech Stack:** Go/Gin, React/TypeScript, existing test tools.

---

## Task 1: Baseline and shared helpers

- [ ] Record current method/path/response behavior for all registered routes and React module callers.
- [ ] Add a failing real-route test for GVA success and business failure semantics.
- [ ] Implement `Response`, `Result`, `Ok`, `Fail`, and `NoAuth` helpers in `server/model/common/response/response.go`.
- [ ] Verify existing migrated routes, old routes, CSV, and health/ping separately.

## Task 2: Project module

- [ ] Add tests for list, create, validation error, permission error, and member operations through registered routes.
- [ ] Update `server/api/v1/project/{project,member,helpers}.go` to use shared response helpers.
- [ ] Update `web/src/api/project.ts`, project response types, and the affected project views/context.
- [ ] Run project/backend and frontend tests and builds; check existing route paths and scope.

## Task 3: Remaining system and cloud modules

- [ ] Migrate role, permission, LDAP, cloud account, and asset JSON routes with their React module functions in batches.
- [ ] Keep CSV download success as CSV; make its JSON error path consistent with the shared response contract.
- [ ] Verify status/code/data/msg for success, business failure, authentication failure, protocol failure, and empty collections.

## Task 4: Remove migration switch

- [ ] Move JWT, RBAC, ProjectScope, and Recovery to one response contract.
- [ ] Remove `EnvelopeForRoutes` and its per-route list from `server/initialize/router.go` after all consumers have switched.
- [ ] Run `go test -count=1 ./...`, `go build ./...`, frontend tests/build, and isolated MySQL HTTP smoke; record old-client compatibility and rollback requirements.

Swagger is deferred per user instruction. Authentication token transport, blacklist, refresh, and Casbin policy conversion are handled in a separate batch after this response chain is stable.
