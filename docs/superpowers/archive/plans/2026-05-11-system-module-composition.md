# System Module Composition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move `system`-domain dependency assembly out of `api` and `middleware` request paths into a single `SystemModule` composition boundary without changing current routes or HTTP behavior.

**Architecture:** Keep the existing `initialize / router / api / service` shell, but introduce a `SystemModule` built at startup time. `router` will mount injected handlers, `middleware` will consume injected capabilities, and `api` handlers will stop reaching into `global` or constructing stores/services on demand. Database-unavailable behavior remains compatible by injecting unavailable dependencies or `nil`-guarded handlers rather than reading `global` during requests.

**Tech Stack:** Go 1.23, Gin, GORM, testify, existing `initialize/router/middleware/api/service` packages

---

## File Structure

### Create

- `server/initialize/system_module.go`
  - Define `SystemModule`, `SystemHandlers`, `SystemCapabilities`, and the builder that wires current `service/system` implementations into handler objects.
- `server/initialize/system_module_test.go`
  - Lock module-building behavior, especially nil-DB compatibility and non-nil public handler assembly.
- `server/api/v1/system/handlers.go`
  - Define handler-facing service interfaces, shared unavailable-response helpers, and the handler bundle shape used by `SystemModule`.

### Modify

- `server/api/v1/system/auth.go`
  - Convert from `Api` methods with on-demand construction to `AuthHandler` methods with injected dependencies.
- `server/api/v1/system/user.go`
  - Convert to `UserHandler` with injected service.
- `server/api/v1/system/role.go`
  - Convert to `RoleHandler` with injected service.
- `server/api/v1/system/ldap.go`
  - Convert to `LDAPHandler` with injected service.
- `server/api/v1/system/audit.go`
  - Convert to `AuditHandler` with injected service.
- `server/api/v1/system/base.go`
  - Convert to `BaseHandler` with injected ping service or existing service-group dependency wrapped at module construction.
- `server/api/v1/system/health.go`
  - Convert to `HealthHandler` with injected health service or existing service-group dependency wrapped at module construction.
- `server/router/system/enter.go`
  - Accept injected handlers instead of instantiating `Api{}`.
- `server/initialize/router.go`
  - Build `SystemModule`, wire `RBAC` middleware with injected capability, and pass module handlers to router registration.
- `server/middleware/casbin_rbac.go`
  - Replace request-time authorizer construction with explicit injected capability.
- `server/router/router_test.go`
  - Preserve public/private route behavior after router injection changes.
- `server/middleware/casbin_rbac_test.go`
  - Lock explicit authorizer injection behavior.

### Test

- `server/initialize/system_module_test.go`
- `server/router/router_test.go`
- `server/middleware/casbin_rbac_test.go`
- Existing focused service tests only if compilation pressure requires them

## Boundary Rules

- `initialize` owns `SystemModule` construction.
- `router` mounts handlers and middleware only; it does not construct services.
- `middleware` consumes injected capabilities only; it does not read `global.OPTIMUS_DB`.
- `api` handlers may map HTTP errors, but may not construct stores/services/clients.
- This plan does **not** split `service/system` into application vs infrastructure packages.

## Task 1: Lock Current Behavior Before Refactor

**Files:**
- Modify: `server/router/router_test.go`
- Modify: `server/middleware/casbin_rbac_test.go`
- Create: `server/initialize/system_module_test.go`

- [ ] **Step 1: Write the failing tests**

```go
func TestNewSystemModuleBuildsPublicHandlersWithoutDatabase(t *testing.T) {
	module := initialize.NewSystemModule(nil, global.OPTIMUS_CONFIG)
	require.NotNil(t, module.Handlers.Auth)
	require.NotNil(t, module.Handlers.Base)
	require.NotNil(t, module.Handlers.Health)
}

func TestCasbinHandlerUsesInjectedAuthorizer(t *testing.T) {
	engine := gin.New()
	engine.GET("/api/v1/system/users",
		middleware.CasbinHandler(fakeMiddlewareRBAC{allowed: true}),
		func(c *gin.Context) { c.Status(http.StatusOK) },
	)
	// assert 200
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./initialize ./middleware ./router -run 'TestNewSystemModuleBuildsPublicHandlersWithoutDatabase|TestCasbinHandlerUsesInjectedAuthorizer|TestRoutersExposeHealthAndPing|TestSystemAdminRoutesRequireJWT' -v
```

Expected:
- `FAIL` because `NewSystemModule` does not exist yet
- `FAIL` because `CasbinHandler` does not accept an injected authorizer yet

- [ ] **Step 3: Add the minimal test scaffolding**

```go
// server/initialize/system_module_test.go
package initialize

func TestNewSystemModuleBuildsPublicHandlersWithoutDatabase(t *testing.T) { /* ... */ }
```

- [ ] **Step 4: Re-run the focused tests and confirm the failure mode stays targeted**

Run the same command as Step 2.

Expected:
- Only the new composition-seam tests fail
- Existing route behavior tests continue to pass

- [ ] **Step 5: Commit**

```bash
git add server/initialize/system_module_test.go server/middleware/casbin_rbac_test.go server/router/router_test.go
git commit -m "Lock system composition refactor behavior before code moves" \
  -m "Constraint: Must preserve current public/private route behavior while moving assembly out of request paths" \
  -m "Rejected: Refactor first without extra tests | too risky for a boundary change" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Keep route and middleware behavior tests green while changing dependency assembly ownership" \
  -m "Tested: go test ./initialize ./middleware ./router -run 'TestNewSystemModuleBuildsPublicHandlersWithoutDatabase|TestCasbinHandlerUsesInjectedAuthorizer|TestRoutersExposeHealthAndPing|TestSystemAdminRoutesRequireJWT' -v" \
  -m "Not-tested: Full server package suite"
```

## Task 2: Introduce `SystemModule` as the Composition Boundary

**Files:**
- Create: `server/initialize/system_module.go`
- Create: `server/initialize/system_module_test.go`
- Modify: `server/api/v1/system/handlers.go`

- [ ] **Step 1: Write the failing module-construction test for nil-DB compatibility**

```go
func TestNewSystemModuleReturnsRBACCapabilityWhenDatabaseExists(t *testing.T) {
	db := newTestGormDB(t)
	module := NewSystemModule(db, global.OPTIMUS_CONFIG)
	require.NotNil(t, module.Capabilities.RBAC)
}
```

- [ ] **Step 2: Run the focused initialize tests**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./initialize -run 'TestNewSystemModule' -v
```

Expected:
- `FAIL` because the builder and module types are not implemented

- [ ] **Step 3: Implement the minimal `SystemModule` builder**

```go
type SystemModule struct {
	Handlers     SystemHandlers
	Capabilities SystemCapabilities
}

type SystemCapabilities struct {
	RBAC middleware.APIAuthorizer
}

func NewSystemModule(db *gorm.DB, cfg config.Config) *SystemModule {
	module := &SystemModule{}
	// build nil-safe handlers for public and private routes
	// build RBAC capability only when db is available
	return module
}
```

- [ ] **Step 4: Re-run initialize tests and confirm they pass**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./initialize -run 'TestNewSystemModule' -v
```

Expected:
- `PASS`

- [ ] **Step 5: Commit**

```bash
git add server/initialize/system_module.go server/initialize/system_module_test.go server/api/v1/system/handlers.go
git commit -m "Recover system assembly ownership in a startup-built module" \
  -m "Constraint: Must keep SystemModule scoped to the system domain and preserve nil-DB compatibility for current route tests" \
  -m "Rejected: Project-wide container introduction | unnecessary scope for the first composition refactor" \
  -m "Confidence: high" \
  -m "Scope-risk: moderate" \
  -m "Directive: Treat SystemModule as a composition boundary, not as a new global registry" \
  -m "Tested: go test ./initialize -run 'TestNewSystemModule' -v" \
  -m "Not-tested: Router and middleware integration"
```

## Task 3: Convert `system` HTTP Handlers to Injected Dependencies

**Files:**
- Create: `server/api/v1/system/handlers.go`
- Modify: `server/api/v1/system/auth.go`
- Modify: `server/api/v1/system/user.go`
- Modify: `server/api/v1/system/role.go`
- Modify: `server/api/v1/system/ldap.go`
- Modify: `server/api/v1/system/audit.go`
- Modify: `server/api/v1/system/base.go`
- Modify: `server/api/v1/system/health.go`

- [ ] **Step 1: Write a failing handler test for one representative injected handler**

```go
func TestAuthHandlerLoginReturnsServiceUnavailableWhenServiceMissing(t *testing.T) {
	handler := system.NewAuthHandler(nil)
	// POST /system/auth/login with valid JSON
	// assert 503 and database unavailable message
}
```

- [ ] **Step 2: Run the focused API test**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./api/v1/system -run 'TestAuthHandlerLoginReturnsServiceUnavailableWhenServiceMissing' -v
```

Expected:
- `FAIL` because the injected handler constructors do not exist yet

- [ ] **Step 3: Implement injected handlers with nil-safe dependencies**

```go
type AuthHandler struct {
	service authService
}

func NewAuthHandler(service authService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Login(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"message": "database is not available"})
		return
	}
	// existing bind + error mapping logic
}
```

Repeat the same pattern for `UserHandler`, `RoleHandler`, `LDAPHandler`, `AuditHandler`, `BaseHandler`, and `HealthHandler`.

- [ ] **Step 4: Re-run focused handler and route regression tests**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./api/v1/system ./router -run 'TestAuthHandlerLoginReturnsServiceUnavailableWhenServiceMissing|TestRoutersExposeHealthAndPing|TestPingRouteReturnsExpectedPayload|TestBasePingRouteReturnsExpectedPayload' -v
```

Expected:
- `PASS`

- [ ] **Step 5: Commit**

```bash
git add server/api/v1/system/handlers.go server/api/v1/system/auth.go server/api/v1/system/user.go server/api/v1/system/role.go server/api/v1/system/ldap.go server/api/v1/system/audit.go server/api/v1/system/base.go server/api/v1/system/health.go
git commit -m "Turn system HTTP handlers into injected protocol adapters" \
  -m "Constraint: Must preserve existing HTTP error mapping while removing request-time service construction" \
  -m "Rejected: Keep one wide Api struct | would hide the same coupling in a new large handler" \
  -m "Confidence: high" \
  -m "Scope-risk: moderate" \
  -m "Directive: Handlers may map HTTP errors, but must not construct stores, services, or clients" \
  -m "Tested: go test ./api/v1/system ./router -run 'TestAuthHandlerLoginReturnsServiceUnavailableWhenServiceMissing|TestRoutersExposeHealthAndPing|TestPingRouteReturnsExpectedPayload|TestBasePingRouteReturnsExpectedPayload' -v" \
  -m "Not-tested: Full middleware integration"
```

## Task 4: Inject `SystemModule` into Router and RBAC Middleware

**Files:**
- Modify: `server/router/system/enter.go`
- Modify: `server/initialize/router.go`
- Modify: `server/middleware/casbin_rbac.go`
- Modify: `server/middleware/casbin_rbac_test.go`
- Modify: `server/router/router_test.go`

- [ ] **Step 1: Write failing integration tests for router + middleware injection**

```go
func TestRoutersUseInjectedRBACCapability(t *testing.T) {
	// build engine with a module whose RBAC authorizer denies the request
	// assert 403 on /api/v1/system/users
}
```

- [ ] **Step 2: Run focused router/middleware tests**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./middleware ./router ./initialize -run 'TestRoutersUseInjectedRBACCapability|TestCasbinHandlerAllowsAuthorizedRoute|TestCasbinHandlerDeniesUnauthorizedRoute|TestAuthMeRouteRequiresJWT|TestSystemAdminRoutesRequireJWT' -v
```

Expected:
- `FAIL` because router and middleware still use internal construction paths

- [ ] **Step 3: Implement explicit injection**

```go
// server/middleware/casbin_rbac.go
type APIAuthorizer interface {
	AuthorizeAPI(ctx context.Context, method, path string, roleIDs []uint) (bool, error)
}

func CasbinHandler(authorizer APIAuthorizer) gin.HandlerFunc { /* ... */ }

// server/router/system/enter.go
func (r *RouterGroup) InitSystemRouter(publicGroup, privateGroup *gin.RouterGroup, handlers system.Handlers) { /* ... */ }

// server/initialize/router.go
module := NewSystemModule(global.OPTIMUS_DB, global.OPTIMUS_CONFIG)
privateGroup.Use(middleware.JWTAuth(), middleware.CasbinHandler(module.Capabilities.RBAC))
routerGroup.System.InitSystemRouter(publicGroup, privateGroup, module.Handlers)
```

- [ ] **Step 4: Re-run focused router/middleware tests**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./middleware ./router ./initialize -run 'TestRoutersUseInjectedRBACCapability|TestCasbinHandlerAllowsAuthorizedRoute|TestCasbinHandlerDeniesUnauthorizedRoute|TestAuthMeRouteRequiresJWT|TestSystemAdminRoutesRequireJWT' -v
```

Expected:
- `PASS`

- [ ] **Step 5: Commit**

```bash
git add server/router/system/enter.go server/initialize/router.go server/middleware/casbin_rbac.go server/middleware/casbin_rbac_test.go server/router/router_test.go
git commit -m "Inject system handlers and RBAC capability from the module boundary" \
  -m "Constraint: Must preserve route paths and current JWT-before-RBAC behavior while removing request-time RBAC assembly" \
  -m "Rejected: Keep a hidden default authorizer constructor in middleware | would preserve assembly in the request path" \
  -m "Confidence: high" \
  -m "Scope-risk: moderate" \
  -m "Directive: Router mounts injected handlers and middleware only; it must not instantiate API shells or authorization services" \
  -m "Tested: go test ./middleware ./router ./initialize -run 'TestRoutersUseInjectedRBACCapability|TestCasbinHandlerAllowsAuthorizedRoute|TestCasbinHandlerDeniesUnauthorizedRoute|TestAuthMeRouteRequiresJWT|TestSystemAdminRoutesRequireJWT' -v" \
  -m "Not-tested: Full server suite"
```

## Task 5: Full Verification and Cleanup

**Files:**
- Modify: any touched files above for final cleanup only

- [ ] **Step 1: Search for leftover request-time assembly**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
grep -R "NewGorm.*Store\\|New.*Service" api/v1 middleware initialize router
```

Expected:
- No `api/v1/system` or `middleware` results that construct `Gorm` stores or `service/system` implementations during requests

- [ ] **Step 2: Run focused and full verification**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./initialize ./middleware ./router ./...
go build ./...
```

Expected:
- All tests pass
- Build succeeds

- [ ] **Step 3: Review changed files for scope discipline**

Checklist:
- No new global variable introduced
- No `SystemModule` expansion beyond system domain
- No `service/system` package split attempted
- No route path or status regression introduced intentionally

- [ ] **Step 4: Commit the completed implementation**

```bash
git add server/api/v1/system server/initialize server/router server/middleware
git commit -m "Move system dependency assembly to a startup-built module" \
  -m "Recover a single system-domain composition boundary so HTTP handlers and RBAC middleware stop constructing business dependencies during requests." \
  -m "Constraint: Must preserve current route behavior and defer deeper service/global/runtime refactors to later work" \
  -m "Rejected: Full application/infrastructure package rewrite | outside the first composition-refactor scope" \
  -m "Confidence: medium" \
  -m "Scope-risk: moderate" \
  -m "Directive: Keep future assembly at startup boundaries; do not reintroduce store/service construction in API handlers or middleware" \
  -m "Tested: go test ./initialize ./middleware ./router ./... && go build ./..." \
  -m "Not-tested: Non-HTTP entrypoints, since this refactor is scoped to current Gin routes"
```

- [ ] **Step 5: Prepare the next refactor handoff**

Write a short follow-up note in the final report covering:
- files changed
- remaining risks
- why issue 2 (`service` mixed with infrastructure) is the next priority

## Review Notes

- Do a local self-review after writing code to confirm the plan stayed inside “problem 1 only”.
- If explicit delegation is requested later, run a plan review or execution review with `@superpowers:subagent-driven-development`.
