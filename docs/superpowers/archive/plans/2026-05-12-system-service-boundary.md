# System Service Boundary Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split `server/service/system` into real `usecase` and `infra` package boundaries without changing current HTTP routes, middleware behavior, or `SystemModule` responsibilities.

**Architecture:** Keep the existing `initialize / router / api / service` shell and move the inner `system` capability code into `server/service/system/usecase`, `server/service/system/infra/persistence`, `server/service/system/infra/identity`, and `server/service/system/infra/directory`. `initialize.SystemModule` becomes the only place that assembles concrete Gorm/JWT/LDAP adapters into usecases, while `api` and `middleware` depend only on the `usecase` boundary or narrow compatibility seams.

**Tech Stack:** Go 1.23, Gin, GORM, testify, `github.com/golang-jwt/jwt/v5`, `github.com/go-ldap/ldap/v3`

---

## File Structure

### Create

- `server/service/system/usecase/auth.go`
  - Move `AuthService`, its DTOs/errors, and auth-owned ports here.
- `server/service/system/usecase/auth_test.go`
  - Preserve the current fake-store auth behavior tests under the new package.
- `server/service/system/usecase/user.go`
  - Move `UserService`, user DTOs/errors, and local ports here.
- `server/service/system/usecase/user_test.go`
  - Preserve the current user fake-store tests under the new package.
- `server/service/system/usecase/role.go`
  - Move `RoleService`, role DTOs/errors, and local ports here.
- `server/service/system/usecase/role_test.go`
  - Preserve the current role fake-store tests under the new package.
- `server/service/system/usecase/ldap.go`
  - Move `LDAPService`, LDAP DTOs/errors, and local ports here.
- `server/service/system/usecase/ldap_test.go`
  - Preserve the current LDAP fake-store tests under the new package.
- `server/service/system/usecase/audit.go`
  - Hold `AuditService`, `AuditMeta`, `AuditEvent`, `AuditLogListResult`, and package-local shared audit types.
- `server/service/system/usecase/support.go`
  - Hold usecase-local constants and helpers needed after extraction, including `defaultAdminUsername`, `defaultUserStatus`, `defaultRoleStatus`, and `resourceID`.
- `server/service/system/usecase/password.go`
  - Hold usecase-local password helpers so extracted auth code does not depend on the root package.
- `server/service/system/usecase/audit_test.go`
  - Preserve or add direct usecase tests for audit list behavior.
- `server/service/system/usecase/rbac.go`
  - Move `RBACService`, `SystemPermissionSummary`, and RBAC ports here.
- `server/service/system/usecase/rbac_test.go`
  - Preserve current RBAC authorization tests under the new package.
- `server/service/system/infra/persistence/auth_store_gorm.go`
  - Move `GormAuthStore` and auth-specific persistence helpers here.
- `server/service/system/infra/persistence/auth_store_gorm_test.go`
  - Keep current persistence-focused auth store tests under the new path.
- `server/service/system/infra/persistence/testdb_test.go`
  - Hold the moved MySQL-backed test database helper copied from the root package test setup.
- `server/service/system/infra/persistence/user_store_gorm_test.go`
  - Add focused user-store persistence tests after extraction.
- `server/service/system/infra/persistence/role_store_gorm_test.go`
  - Add focused role-store persistence tests after extraction.
- `server/service/system/infra/persistence/ldap_store_gorm_test.go`
  - Add focused LDAP-store persistence tests after extraction.
- `server/service/system/infra/persistence/audit_store_gorm_test.go`
  - Add focused audit-store persistence tests after extraction.
- `server/service/system/infra/persistence/rbac_store_gorm_test.go`
  - Add focused RBAC-store persistence tests after extraction.
- `server/service/system/infra/persistence/user_store_gorm.go`
  - Move `GormUserStore` here.
- `server/service/system/infra/persistence/role_store_gorm.go`
  - Move `GormRoleStore` here.
- `server/service/system/infra/persistence/ldap_store_gorm.go`
  - Move `GormLDAPStore` here.
- `server/service/system/infra/persistence/audit_store_gorm.go`
  - Move `GormAuditStore` here.
- `server/service/system/infra/persistence/rbac_store_gorm.go`
  - Move `GormRBACStore` here.
- `server/service/system/infra/identity/jwt.go`
  - Move the real JWT issue/parse implementation here and expose the concrete token-issuer/parser adapter shape chosen for the new usecase ports.
- `server/service/system/infra/identity/jwt_test.go`
  - Preserve the JWT issue/parse tests under the new path.
- `server/service/system/infra/directory/ldap_client.go`
  - Move `GoLDAPClient` and LDAP helper functions here.
- `server/service/system/infra/directory/ldap_client_test.go`
  - Add focused tests for helper behavior that should stay below the usecase boundary.

### Modify

- `server/initialize/system_module.go`
  - Wire new `infra -> usecase` dependencies instead of root-package mixed services.
- `server/initialize/system_module_test.go`
  - Keep nil-DB and capability wiring behavior green after the package split.
- `server/initialize/router_test.go`
  - Migrate fake auth-store/test wiring to `usecase` types while keeping JWT issuance on the approved root compatibility seam.
- `server/api/v1/system/handlers.go`
  - Switch handler dependency types from `server/service/system` to `server/service/system/usecase`.
- `server/api/v1/system/auth.go`
  - Bind/authenticate against `usecase.AuthService` types and errors.
- `server/api/v1/system/user.go`
  - Bind/update against `usecase.UserService` types and errors.
- `server/api/v1/system/role.go`
  - Bind/update against `usecase.RoleService` types and errors, while intentionally leaving `PermissionTree` on the bounded `DefaultSystemPermissions()` seam from `seed.go` in this round.
- `server/api/v1/system/ldap.go`
  - Bind/test against `usecase.LDAPService` types and errors.
- `server/api/v1/system/audit.go`
  - Read audit results from `usecase.AuditService`.
- `server/middleware/jwt.go`
  - Keep middleware behavior unchanged while routing parse calls only through the narrow root compatibility seam preserved in `server/service/system/jwt.go`.
- `server/middleware/jwt_test.go`
  - Keep current middleware token behavior green after JWT implementation moves.
- `server/service/system/jwt.go`
  - Reduce to a thin compatibility wrapper around `infra/identity`.
- `server/service/system/auth.go`
  - Remove real auth usecase logic from the root package; keep only compatibility glue if still required during migration.
- `server/service/system/user.go`
  - Remove real user usecase/store logic from the root package.
- `server/service/system/role.go`
  - Remove real role usecase/store logic from the root package.
- `server/service/system/ldap.go`
  - Remove real LDAP usecase/store logic from the root package.
- `server/service/system/audit.go`
  - Remove real audit usecase/store logic from the root package.
- `server/service/system/rbac.go`
  - Remove real RBAC usecase/store logic from the root package.
- `server/service/system/auth_store.go`
  - Remove real Gorm auth store implementation from the root package.
- `server/service/system/ldap_client.go`
  - Remove real LDAP client implementation from the root package.
- `server/service/system/auth_test.go`
  - Retire after its assertions are ported to `server/service/system/usecase/auth_test.go`.
- `server/service/system/user_test.go`
  - Retire after its assertions are ported to `server/service/system/usecase/user_test.go`.
- `server/service/system/role_test.go`
  - Retire after its assertions are ported to `server/service/system/usecase/role_test.go`.
- `server/service/system/ldap_test.go`
  - Retire after its assertions are ported to `server/service/system/usecase/ldap_test.go`.
- `server/service/system/rbac_test.go`
  - Retire after its assertions are ported to `server/service/system/usecase/rbac_test.go`.
- `server/service/system/jwt_test.go`
  - Retire after its assertions are ported to `server/service/system/infra/identity/jwt_test.go`.

### Test

- `server/service/system/usecase/auth_test.go`
- `server/service/system/usecase/user_test.go`
- `server/service/system/usecase/role_test.go`
- `server/service/system/usecase/ldap_test.go`
- `server/service/system/usecase/audit_test.go`
- `server/service/system/usecase/rbac_test.go`
- `server/service/system/infra/persistence/auth_store_gorm_test.go`
- `server/service/system/infra/persistence/user_store_gorm_test.go`
- `server/service/system/infra/persistence/role_store_gorm_test.go`
- `server/service/system/infra/persistence/ldap_store_gorm_test.go`
- `server/service/system/infra/persistence/audit_store_gorm_test.go`
- `server/service/system/infra/persistence/rbac_store_gorm_test.go`
- `server/service/system/infra/identity/jwt_test.go`
- `server/service/system/infra/directory/ldap_client_test.go`
- `server/initialize/system_module_test.go`
- `server/api/v1/system/handlers_test.go`
- `server/middleware/jwt_test.go`
- `server/router/router_test.go`

## Boundary Rules

- `usecase` owns stable DTOs, semantic errors, and local ports.
- Shared types needed by multiple usecases, such as `AuditMeta` and `AuditEvent`, may live inside the `usecase` package itself, not in a separate `shared` contract package.
- `usecase` may keep local status/admin constants for business logic, but `infra/persistence` should use its own package-local stable literals like `"enabled"` rather than importing unexported usecase constants.
- `infra/persistence` may import GORM and `systemModel`, but not `gin` or HTTP concerns.
- `infra/identity` and `infra/directory` may import config and third-party libraries, but the exposed surface back to `usecase` must stay narrow.
- `api` handlers may depend on concrete `usecase` services or one-to-one thin interfaces only; they must not reach into `infra`.
- `server/service/system` is a shrinking compatibility seam, not a new facade layer.

## Temporary Compatibility Matrix

- Root exports that may remain temporarily in `server/service/system` during this plan:
  - `IssueJWT` and `ParseJWT`, as thin forwarding wrappers for `middleware/jwt.go` and JWT-oriented tests
  - `DefaultSystemPermissions()`, only for the permissions-tree endpoint backed by `seed.go`, which the spec keeps out of the main refactor scope
  - `HashPassword` and `ComparePassword`, only for `seed.go` / `seed_test.go`, which stay outside the main boundary split in this round
- Callers that must migrate off root exports in this plan:
  - `server/initialize/system_module.go`
  - `server/api/v1/system/auth.go`
  - `server/api/v1/system/user.go`
  - `server/api/v1/system/role.go`
  - `server/api/v1/system/ldap.go`
  - `server/api/v1/system/audit.go`
  - `server/initialize/router_test.go` for `NewAuthService`, `TokenSubject`, `LDAPUser`, `SystemRoleSummary`, and `AuditEvent`
- Non-goal:
  - Do not preserve broad root aliases for usecase constructors or DTOs just to avoid touching tests

## Task 1: Lock the New Boundary With Red Tests

**Files:**
- Create: `server/service/system/usecase/auth_test.go`
- Create: `server/service/system/usecase/user_test.go`
- Create: `server/service/system/usecase/role_test.go`
- Create: `server/service/system/usecase/ldap_test.go`
- Create: `server/service/system/usecase/rbac_test.go`
- Create: `server/service/system/infra/identity/jwt_test.go`
- Modify: `server/initialize/system_module_test.go`
- Modify: `server/middleware/jwt_test.go`

- [ ] **Step 1: Create the new package directories and copy one representative test per target boundary into those paths**

```go
func TestAuthServiceLoginAdminSuccess(t *testing.T) {
	store := newFakeAuthStore(...)
	service := usecase.NewAuthService(store, fakeTokenIssuer{}, fakeLDAPAuthenticator{})
	// assert token, roles, permissions, audit side effects
}

func TestIssueAndParseJWT(t *testing.T) {
	token, err := identity.JWTIssuer{Config: testAuthJWTConfig()}.IssueToken(...)
	claims, err := identity.ParseJWT(token, cfg)
	// assert round-trip claims
}
```

- [ ] **Step 2: Run the focused boundary tests and confirm they fail on missing constructors, DTOs, ports, or adapter symbols**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system/usecase ./service/system/infra/identity ./initialize ./middleware -run 'TestAuthServiceLoginAdminSuccess|TestIssueAndParseJWT|TestNewSystemModuleBuildsPublicHandlersWithoutDatabase|TestJWTAuthStoresClaimsForValidToken' -v
```

Expected:
- `FAIL` with missing-symbol or missing-implementation errors under `usecase` / `infra/identity`

- [ ] **Step 3: Add only the minimal test-only scaffolding needed to keep the failure on missing implementation symbols**

```go
type fakeTokenIssuer struct{}

func (f fakeTokenIssuer) IssueToken(subject TokenSubject) (string, error) {
	return "", nil
}
```

- [ ] **Step 4: Re-run the focused tests and confirm the failure stays on missing constructors, DTOs, or ports, not malformed scaffolding**

Run the same command as Step 2.

Expected:
- `FAIL` on missing constructors, DTOs, or adapter functions

- [ ] **Step 5: Commit**

```bash
git add server/service/system/usecase server/service/system/infra/identity server/initialize/system_module_test.go server/middleware/jwt_test.go
git commit -m "Lock the system boundary split with package-level red tests" \
  -m "Constraint: Must preserve current HTTP and middleware behavior while moving implementations into subpackages" \
  -m "Rejected: Refactor first without new package-boundary tests | too easy to preserve compilation while losing the intended seam" \
  -m "Confidence: high" \
  -m "Scope-risk: narrow" \
  -m "Directive: Keep tests proving the new package boundaries, not just the old root package behavior" \
  -m "Tested: go test ./service/system/usecase ./service/system/infra/identity ./initialize ./middleware -run 'TestAuthServiceLoginAdminSuccess|TestIssueAndParseJWT|TestNewSystemModuleBuildsPublicHandlersWithoutDatabase|TestJWTAuthStoresClaimsForValidToken' -v" \
  -m "Not-tested: Full server package suite"
```

## Task 2: Extract the `usecase` Package and Shared In-Package Audit Types

**Files:**
- Create: `server/service/system/usecase/audit.go`
- Create: `server/service/system/usecase/support.go`
- Create: `server/service/system/usecase/password.go`
- Create: `server/service/system/usecase/audit_test.go`
- Create: `server/service/system/usecase/auth.go`
- Create: `server/service/system/usecase/user.go`
- Create: `server/service/system/usecase/role.go`
- Create: `server/service/system/usecase/ldap.go`
- Create: `server/service/system/usecase/rbac.go`
- Create: `server/service/system/usecase/auth_test.go`
- Create: `server/service/system/usecase/user_test.go`
- Create: `server/service/system/usecase/role_test.go`
- Create: `server/service/system/usecase/ldap_test.go`
- Create: `server/service/system/usecase/rbac_test.go`

- [ ] **Step 1: Move `AuditMeta`, `AuditEvent`, `AuditLogListResult`, and `AuditService` into `usecase/audit.go`, and add a direct audit usecase test file**

```go
package usecase

type AuditMeta struct {
	IP        string
	UserAgent string
}

type AuditEvent struct {
	ActorUserID  uint
	Action       string
	ResourceType string
	ResourceID   string
	Result       string
	IP           string
	UserAgent    string
	Detail       string
}
```

- [ ] **Step 2: Create `usecase/support.go` for extraction-owned constants/helpers, and `usecase/password.go` for the auth-local hashing helpers**

```go
package usecase

const (
	defaultAdminUsername = "admin"
	defaultUserStatus    = "enabled"
	defaultRoleStatus    = "enabled"
)

func resourceID(id uint) string { ... }
```

```go
package usecase

func HashPassword(password string) (string, error) { ... }
func ComparePassword(hash string, password string) bool { ... }
```

- [ ] **Step 3: Run the smallest audit-focused usecase test**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system/usecase -run 'TestAuditService' -v
```

Expected:
- `FAIL` if `AuditService` or its types are not wired correctly yet

- [ ] **Step 4: Move auth, user, role, LDAP, and RBAC usecase logic into the new package without importing GORM or concrete clients**

```go
type AuthService struct {
	store       AuthStore
	tokenIssuer TokenIssuer
	ldap        LDAPAuthenticator
	now         func() time.Time
}

type UserService struct {
	store UserStore
}
```

- [ ] **Step 5: Port the copied usecase tests and make them pass in the new package**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system/usecase -run 'TestAuthService|TestUserService|TestRoleService|TestLDAPService|TestRBACService|TestAuditService' -v
```

Expected:
- `PASS`
- `server/api/v1/system/handlers_test.go` remains green without reintroducing root-package service dependencies
- Root `seed.go` and `seed_test.go` remain unaffected because `server/service/system/password.go` is intentionally retained in this round

- [ ] **Step 6: Commit**

```bash
git add server/service/system/usecase
git commit -m "Move system business logic into a transport-agnostic usecase package" \
  -m "Constraint: Usecases must own DTOs and semantic errors without importing GORM, Gin, or concrete JWT/LDAP clients" \
  -m "Rejected: Separate shared contract package | would create a premature abstraction layer inside system" \
  -m "Confidence: high" \
  -m "Scope-risk: moderate" \
  -m "Directive: Keep shared audit and summary types inside usecase unless a later spec explicitly extracts policy or shared contracts" \
  -m "Tested: go test ./service/system/usecase -run 'TestAuthService|TestUserService|TestRoleService|TestLDAPService|TestRBACService|TestAuditService' -v" \
  -m "Not-tested: GORM-backed adapter behavior"
```

## Task 3: Extract JWT and LDAP Adapters Out of the Root Package

**Files:**
- Create: `server/service/system/infra/identity/jwt.go`
- Create: `server/service/system/infra/identity/jwt_test.go`
- Create: `server/service/system/infra/directory/ldap_client.go`
- Create: `server/service/system/infra/directory/ldap_client_test.go`
- Modify: `server/service/system/jwt.go`
- Modify: `server/middleware/jwt.go`
- Modify: `server/middleware/jwt_test.go`

- [ ] **Step 1: Move the real JWT implementation into `infra/identity/jwt.go`**

```go
package identity

type JWTIssuer struct {
	Config config.JWT
	Now    func() time.Time
}

func (i JWTIssuer) IssueToken(subject usecase.TokenSubject) (string, error) { ... }
func ParseJWT(tokenString string, cfg config.JWT) (*utils.CustomClaims, error) { ... }
```

- [ ] **Step 2: Re-run JWT adapter and middleware tests while the root wrapper is still missing**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system/infra/identity ./middleware -run 'TestIssueAndParseJWT|TestJWTAuth' -v
```

Expected:
- Adapter tests may pass
- Middleware tests may still fail until the root compatibility seam is restored

- [ ] **Step 3: Move `GoLDAPClient` into `infra/directory/ldap_client.go` and add focused helper tests before the verification run**

```go
package directory

type GoLDAPClient struct {
	Timeout time.Duration
}

func TestReplaceLDAPUsername(t *testing.T) { ... }
func TestFirstLDAPAttr(t *testing.T) { ... }
```

- [ ] **Step 4: Reintroduce only a thin root JWT wrapper and keep middleware on that wrapper**

```go
package system

func IssueJWT(subject usecase.TokenSubject, cfg config.JWT, now time.Time) (string, error) {
	return identity.JWTIssuer{Config: cfg, Now: func() time.Time { return now }}.IssueToken(subject)
}
```

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system/infra/identity ./service/system/infra/directory ./middleware -run 'TestIssueAndParseJWT|TestJWTAuth|TestReplaceLDAPUsername|TestFirstLDAPAttr' -v
```

Expected:
- `PASS`

- [ ] **Step 5: Commit**

```bash
git add server/service/system/infra/identity server/service/system/infra/directory server/service/system/jwt.go server/middleware/jwt.go server/middleware/jwt_test.go
git commit -m "Move JWT and LDAP adapters behind explicit infrastructure packages" \
  -m "Constraint: JWT middleware behavior must remain compatible while the real implementation leaves the root package" \
  -m "Rejected: Let middleware import infra/identity directly | couples transport code to a concrete adapter package" \
  -m "Confidence: high" \
  -m "Scope-risk: moderate" \
  -m "Directive: Keep root jwt.go as a forwarding seam only; do not let new code grow there" \
  -m "Tested: go test ./service/system/infra/identity ./service/system/infra/directory ./middleware -run 'TestIssueAndParseJWT|TestJWTAuth|TestReplaceLDAPUsername|TestFirstLDAPAttr' -v" \
  -m "Not-tested: Full auth and ldap package integration"
```

## Task 4: Extract GORM Persistence Adapters

**Files:**
- Create: `server/service/system/infra/persistence/auth_store_gorm.go`
- Create: `server/service/system/infra/persistence/auth_store_gorm_test.go`
- Create: `server/service/system/infra/persistence/testdb_test.go`
- Create: `server/service/system/infra/persistence/user_store_gorm.go`
- Create: `server/service/system/infra/persistence/user_store_gorm_test.go`
- Create: `server/service/system/infra/persistence/role_store_gorm.go`
- Create: `server/service/system/infra/persistence/role_store_gorm_test.go`
- Create: `server/service/system/infra/persistence/ldap_store_gorm.go`
- Create: `server/service/system/infra/persistence/ldap_store_gorm_test.go`
- Create: `server/service/system/infra/persistence/audit_store_gorm.go`
- Create: `server/service/system/infra/persistence/audit_store_gorm_test.go`
- Create: `server/service/system/infra/persistence/rbac_store_gorm.go`
- Create: `server/service/system/infra/persistence/rbac_store_gorm_test.go`
- Modify: `server/service/system/auth_store_test.go`

- [ ] **Step 1: Move `GormAuthStore` first and have it satisfy the new `usecase.AuthStore` port**

```go
package persistence

type GormAuthStore struct {
	db *gorm.DB
}
```

- [ ] **Step 2: Run the focused auth store tests and fix compile errors at the new import boundary**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system/infra/persistence -run 'TestMenuPermissionCodesForRolesOrdersDistinctCodes|TestGormAuthStoreSyncLDAPUserCreatesDefaultRole' -v
```

Expected:
- `FAIL` until the moved store compiles against `usecase` types

- [ ] **Step 3: Move user, role, LDAP, audit, and RBAC GORM stores into the persistence package**

```go
const enabledUserStatus = "enabled"
const enabledRoleStatus = "enabled"

type GormUserStore struct { db *gorm.DB }
type GormRoleStore struct { db *gorm.DB }
type GormLDAPStore struct { db *gorm.DB }
type GormAuditStore struct { db *gorm.DB }
type GormRBACStore struct { db *gorm.DB }
```

- [ ] **Step 4: Add `infra/persistence/testdb_test.go` and focused persistence tests for the moved stores**

```go
func openPersistenceTestDB(t *testing.T) *gorm.DB { ... }

func TestGormAuthStoreSyncLDAPUserCreatesDefaultRole(t *testing.T) { ... }
func TestGormUserStoreWouldHaveAdminCapableUserAfterStatus(t *testing.T) { ... }
func TestGormRoleStoreReplaceRolePermissions(t *testing.T) { ... }
func TestGormLDAPStoreGetLDAPConfigReturnsEmptyWhenMissing(t *testing.T) { ... }
func TestGormAuditStoreListAuditLogsOrdersNewestFirst(t *testing.T) { ... }
func TestGormRBACStoreFindAPIPermissionNotFound(t *testing.T) { ... }
```

- [ ] **Step 5: Re-run the persistence and usecase tests together**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system/usecase ./service/system/infra/persistence -run 'TestAuthService|TestUserService|TestRoleService|TestLDAPService|TestRBACService|TestMenuPermissionCodesForRolesOrdersDistinctCodes|TestGormAuthStoreSyncLDAPUserCreatesDefaultRole|TestGormUserStore|TestGormRoleStore|TestGormLDAPStore|TestGormAuditStore|TestGormRBACStore' -v
```

Expected:
- `PASS`

- [ ] **Step 6: Commit**

```bash
git add server/service/system/infra/persistence server/service/system/auth_store_test.go
git commit -m "Separate GORM-backed system stores from usecase code" \
  -m "Constraint: Persistence adapters may depend on GORM and models but must not absorb business-rule decisions from usecases" \
  -m "Rejected: Keep auth store in the root package as a special case | would leave the main boundary half-complete" \
  -m "Confidence: high" \
  -m "Scope-risk: moderate" \
  -m "Directive: If a store needs a usecase type, import only that narrow type rather than rebuilding a broad shared layer" \
  -m "Tested: go test ./service/system/usecase ./service/system/infra/persistence -run 'TestAuthService|TestUserService|TestRoleService|TestLDAPService|TestRBACService|TestMenuPermissionCodesForRolesOrdersDistinctCodes|TestGormAuthStoreSyncLDAPUserCreatesDefaultRole|TestGormUserStore|TestGormRoleStore|TestGormLDAPStore|TestGormAuditStore|TestGormRBACStore' -v" \
  -m "Not-tested: Full server package suite"
```

## Task 5: Rewire `SystemModule` and HTTP Handlers to the New Boundary

**Files:**
- Modify: `server/initialize/system_module.go`
- Modify: `server/initialize/system_module_test.go`
- Modify: `server/initialize/router_test.go`
- Modify: `server/api/v1/system/handlers.go`
- Modify: `server/api/v1/system/auth.go`
- Modify: `server/api/v1/system/user.go`
- Modify: `server/api/v1/system/role.go`
- Modify: `server/api/v1/system/ldap.go`
- Modify: `server/api/v1/system/audit.go`

- [ ] **Step 1: Update `SystemModule` to construct `persistence`, `identity`, and `directory` adapters and pass them into `usecase` constructors**

```go
func newSystemAuthService(db *gorm.DB, cfg config.Config) *usecase.AuthService {
	if db == nil {
		return nil
	}
	return usecase.NewAuthService(
		persistence.NewGormAuthStore(db),
		identity.JWTIssuer{Config: cfg.JWT},
		directory.NewGoLDAPClient(),
	)
}
```

- [ ] **Step 2: Run the initialize tests and confirm wiring failures surface before handler import changes**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./initialize -run 'TestNewSystemModule' -v
```

Expected:
- `FAIL` on stale root-package imports or constructor signatures

- [ ] **Step 3: Change handlers, request DTO bindings, and `initialize/router_test.go` fakes to import `server/service/system/usecase`**

```go
import systemUsecase "optimus/server/service/system/usecase"

type AuthHandler struct {
	service *systemUsecase.AuthService
}
```

```go
func (fakeRouterAuthStore) SyncLDAPUser(context.Context, systemUsecase.LDAPUser) (systemModel.SystemUser, error) {
	return systemModel.SystemUser{}, errors.New("not implemented")
}
```

Also update the rest of `server/initialize/router_test.go` in the same step:

- replace `systemService.NewAuthService(...)` with `systemUsecase.NewAuthService(...)`
- replace root `TokenSubject` references with `systemUsecase.TokenSubject`
- replace root `SystemRoleSummary` references with `systemUsecase.SystemRoleSummary`
- replace root `AuditEvent` references with `systemUsecase.AuditEvent`
- keep JWT issuance itself on the approved root compatibility seam (`systemService.IssueJWT`)

- [ ] **Step 4: Re-run initialize, API, and router tests**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./initialize ./api/v1/system ./router -run 'TestNewSystemModule|TestAuthHandlerLoginReturnsServiceUnavailableWhenServiceMissing|TestRoutersExposeHealthAndPing|TestSystemAdminRoutesRequireJWT' -v
```

Expected:
- `PASS`

- [ ] **Step 5: Commit**

```bash
git add server/initialize/system_module.go server/initialize/system_module_test.go server/initialize/router_test.go server/api/v1/system/handlers.go server/api/v1/system/auth.go server/api/v1/system/user.go server/api/v1/system/role.go server/api/v1/system/ldap.go server/api/v1/system/audit.go
git commit -m "Wire the system module and handlers to the new usecase and infra packages" \
  -m "Constraint: API handlers must stop at usecase services or one-to-one thin interfaces and must not reach into concrete adapters" \
  -m "Rejected: Leave SystemModule building root-package services | would preserve the old mixed boundary behind a new package tree" \
  -m "Confidence: high" \
  -m "Scope-risk: moderate" \
  -m "Directive: Keep all request-time construction out of handlers; initialize owns assembly" \
  -m "Tested: go test ./initialize ./api/v1/system ./router -run 'TestNewSystemModule|TestAuthHandlerLoginReturnsServiceUnavailableWhenServiceMissing|TestRoutersExposeHealthAndPing|TestSystemAdminRoutesRequireJWT' -v" \
  -m "Not-tested: Middleware + router + usecase full package sweep"
```

## Task 6: Thin the Root Package and Finish Verification

**Files:**
- Modify: `server/service/system/auth.go`
- Modify: `server/service/system/user.go`
- Modify: `server/service/system/role.go`
- Modify: `server/service/system/ldap.go`
- Modify: `server/service/system/audit.go`
- Modify: `server/service/system/rbac.go`
- Modify: `server/service/system/auth_store.go`
- Modify: `server/service/system/ldap_client.go`
- Modify: `server/service/system/jwt.go`
- Modify: `server/service/system/auth_test.go`
- Modify: `server/service/system/user_test.go`
- Modify: `server/service/system/role_test.go`
- Modify: `server/service/system/ldap_test.go`
- Modify: `server/service/system/rbac_test.go`
- Modify: `server/service/system/jwt_test.go`
- Modify: `server/router/router_test.go`
- Modify: `server/middleware/casbin_rbac_test.go`
- Modify: `server/middleware/jwt_test.go`

- [ ] **Step 1: Remove real business and adapter logic from the root package, and retire the migrated root usecase/JWT tests**

```go
// server/service/system/jwt.go
package system

func IssueJWT(subject usecase.TokenSubject, cfg config.JWT, now time.Time) (string, error) {
	return identity.JWTIssuer{Config: cfg, Now: func() time.Time { return now }}.IssueToken(subject)
}

func ParseJWT(tokenString string, cfg config.JWT) (*utils.CustomClaims, error) {
	return identity.ParseJWT(tokenString, cfg)
}
```

Allowed remainder after this cleanup:

- `server/service/system/jwt.go` thin forwarding only
- `server/service/system/seed.go` and its `DefaultSystemPermissions()` helper for the permissions-tree endpoint only
- `server/service/system/password.go` and its hashing helpers for `seed.go` / `seed_test.go` only
- `server/service/system/health.go` and `system.Service` status payloads for the existing ping/health endpoints, which are explicitly out of scope for this refactor

- [ ] **Step 2: Run a repo-wide grep check to confirm no request-path code constructs root-package GORM stores or LDAP/JWT adapters**

Run:

```bash
cd /home/lilin/trae-project/Optimus
grep -RIn "NewGormAuthStore\|NewGormUserStore\|NewGormRoleStore\|NewGormLDAPStore\|NewGormAuditStore\|NewGormRBACStore\|NewGoLDAPClient\|IssueJWT\|ParseJWT" server/api server/middleware server/initialize
```

Expected:
- Only `initialize/system_module.go`, the intended middleware compatibility seam, and tests should remain

- [ ] **Step 3: Run the full focused verification sweep**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system/usecase/... ./service/system/infra/... ./initialize ./middleware ./router ./api/v1/system -v
```

Expected:
- `PASS`

- [ ] **Step 4: Run the broader `service/system` and router-adjacent regression sweep**

Run:

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/system ./service/system/usecase/... ./service/system/infra/... ./initialize ./middleware ./router -v
```

Expected:
- `PASS`

- [ ] **Step 5: Commit**

```bash
git add server/service/system server/router/router_test.go server/middleware/casbin_rbac_test.go server/middleware/jwt_test.go
git commit -m "Finish the system boundary split by shrinking the legacy root package" \
  -m "Constraint: The root package may remain only as a narrow compatibility seam for unavoidable callers" \
  -m "Rejected: Keep long-lived constructor forwarding for all services | would leave server/service/system as the de facto facade" \
  -m "Confidence: medium" \
  -m "Scope-risk: moderate" \
  -m "Directive: Delete compatibility wrappers as soon as their last caller migrates; do not add new logic to the root package" \
  -m "Tested: go test ./service/system/usecase/... ./service/system/infra/... ./initialize ./middleware ./router ./api/v1/system -v" \
  -m "Tested: go test ./service/system ./service/system/usecase/... ./service/system/infra/... ./initialize ./middleware ./router -v" \
  -m "Not-tested: Full server/all-packages sweep including unrelated startup and seed flows"
```

## Final Verification Checklist

- [ ] `server/service/system/usecase` contains the real business logic and semantic errors.
- [ ] `server/service/system/infra/persistence` contains all concrete GORM store implementations.
- [ ] `server/service/system/infra/identity` contains the real JWT implementation.
- [ ] `server/service/system/infra/directory` contains the real LDAP client implementation.
- [ ] `server/initialize/system_module.go` is the only production assembly point for concrete system-domain adapters.
- [ ] `server/api/v1/system` depends only on `usecase` services or thin one-to-one interfaces.
- [ ] `server/middleware/jwt.go` still behaves the same, even if it temporarily uses a root compatibility wrapper.
- [ ] Focused package tests pass for `./service/system/usecase/...`, `./service/system/infra/...`, `./initialize`, `./middleware`, `./router`, and `./api/v1/system`.
