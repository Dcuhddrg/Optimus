# Optimus Backend Skeleton Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## Status Snapshot

As of 2026-04-23, the implementation is ahead of the unchecked boxes below:

- Task 1 completed
- Task 2 completed
- Task 3 completed
- Task 4 completed
- Task 5 completed
- Task 6 completed for the current lightweight router shape
- Task 7 completed with lightweight response helpers; `/api/v1/health`, `/api/v1/ping`, and `/api/v1/base/ping` are wired through `router -> api -> service`
- Task 8 completed: `go test ./...` and `go build ./...` pass; README startup notes are present; live smoke checks passed for `/api/v1/health`, `/api/v1/ping`, and `/api/v1/base/ping` on a temporary port override because port `8888` was already occupied in the environment

**Goal:** Build the first runnable Go backend skeleton under `server/` using a gin-vue-admin-inspired structure with MySQL-first configuration, pluggable Redis/JWT/Casbin wiring, and minimal system health endpoints.

**Architecture:** The implementation keeps `server/main.go` as the only entrypoint, `server/core` as the runtime core, `server/initialize` as the wiring layer, and `server/router` as the route registration layer. The first iteration stays intentionally small: only startup plumbing, config loading, global state, middleware placeholders, and two public system endpoints are implemented, while auth and permission flows remain as extension points.

**Tech Stack:** Go 1.23, Gin 1.10, GORM 1.25, Viper 1.19, Zap 1.27, Redis go-redis/v9, Casbin v2, golang-jwt/jwt/v5, stretchr/testify

---

## File Map

### Create

- `server/go.mod` — backend module definition and direct dependencies
- `server/go.sum` — module dependency lockfile
- `server/main.go` — program entrypoint
- `server/config.yaml` — default local configuration
- `server/config/config.go` — root config aggregation
- `server/config/system.go` — system config structure
- `server/config/zap.go` — zap config structure
- `server/config/gorm_mysql.go` — MySQL config structure
- `server/config/redis.go` — Redis config structure
- `server/config/jwt.go` — JWT config structure
- `server/config/casbin.go` — Casbin config structure
- `server/core/viper.go` — config loading and env override setup
- `server/core/zap.go` — logger initialization
- `server/core/server.go` — HTTP server boot logic
- `server/core/banner.go` — startup banner rendering
- `server/global/global.go` — global singleton objects
- `server/global/version.go` — application version constant
- `server/initialize/init.go` — non-infrastructure init helpers
- `server/initialize/gorm.go` — GORM initialization dispatcher
- `server/initialize/gorm_mysql.go` — MySQL dialector and logger wiring
- `server/initialize/redis.go` — Redis initialization
- `server/initialize/casbin.go` — Casbin initialization
- `server/initialize/router.go` — Gin engine creation and route mounting
- `server/initialize/validator.go` — validator registration placeholder
- `server/router/enter.go` — route group aggregator
- `server/router/system/enter.go` — system router group definition
- `server/api/v1/system/health.go` — health handler
- `server/api/v1/system/base.go` — ping handler
- `server/service/enter.go` — service group aggregator
- `server/service/system/health.go` — system health service
- `server/middleware/logger.go` — request logger middleware
- `server/middleware/recovery.go` — panic recovery middleware
- `server/middleware/jwt.go` — JWT auth placeholder middleware
- `server/middleware/casbin_rbac.go` — Casbin middleware placeholder
- `server/model/common/response.go` — common response payloads
- `server/utils/response.go` — response helpers
- `server/utils/claims.go` — JWT claims helper struct
- `server/docs/.gitkeep` — docs directory placeholder
- `server/router/router_test.go` — route registration tests
- `server/core/viper_test.go` — config loading tests

### Modify

- `README.md` — update project overview and backend startup note

## Task 1: Initialize the backend Go module

**Files:**
- Create: `server/go.mod`
- Create: `server/main.go`
- Create: `server/docs/.gitkeep`
- Modify: `README.md`

- [ ] **Step 1: Write the failing module smoke test**

Create `server/main_test.go` with a minimal test that imports the module root package path expectations indirectly by compiling `main.go`.

```go
package main

import "testing"

func TestMainPackageBuilds(t *testing.T) {}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL because `server/go.mod` and implementation files do not exist yet.

- [ ] **Step 3: Write minimal implementation**

Create `server/go.mod` with:

```go
module optimus/server

go 1.23
```

Create `server/main.go` with a `main()` function that calls `initializeSystem()` and `core.RunServer()`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: package discovery succeeds once stubs exist, though downstream tests may still fail until later tasks land.

- [ ] **Step 5: Commit**

```bash
git add README.md server/go.mod server/main.go server/docs/.gitkeep server/main_test.go
git commit -m "Establish the Go backend entrypoint"
```

## Task 2: Implement configuration structures and default config

**Files:**
- Create: `server/config.yaml`
- Create: `server/config/config.go`
- Create: `server/config/system.go`
- Create: `server/config/zap.go`
- Create: `server/config/gorm_mysql.go`
- Create: `server/config/redis.go`
- Create: `server/config/jwt.go`
- Create: `server/config/casbin.go`
- Create: `server/core/viper.go`
- Create: `server/global/global.go`
- Create: `server/global/version.go`
- Test: `server/core/viper_test.go`

- [ ] **Step 1: Write the failing config loading test**

Create `server/core/viper_test.go`:

```go
func TestViperLoadsConfigFileAndEnvOverride(t *testing.T) {
    t.Setenv("OPTIMUS_SYSTEM_ADDR", "19091")
    cfg, err := LoadConfigForTest("../config.yaml")
    require.NoError(t, err)
    require.Equal(t, 19091, cfg.System.Addr)
    require.Equal(t, "mysql", cfg.System.DBType)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core -run TestViperLoadsConfigFileAndEnvOverride -v`
Expected: FAIL because config structures and loader helpers are missing.

- [ ] **Step 3: Write minimal implementation**

Implement:

- `config.Config` aggregating `System`, `Zap`, `Mysql`, `Redis`, `JWT`, `Casbin`
- `core.Viper()` and a test helper that reads `server/config.yaml`
- environment binding using `OPTIMUS_` prefix and key replacer
- `global.GVA_CONFIG`-style singleton fields renamed for Optimus

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./core -run TestViperLoadsConfigFileAndEnvOverride -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/config.yaml server/config/*.go server/core/viper.go server/core/viper_test.go server/global/*.go
git commit -m "Load backend config from file and environment"
```

## Task 3: Implement core logging and startup wiring

**Files:**
- Create: `server/core/zap.go`
- Create: `server/core/banner.go`
- Create: `server/core/server.go`
- Create: `server/initialize/init.go`
- Modify: `server/main.go`

- [ ] **Step 1: Write the failing startup unit test**

Add to `server/core/server_test.go`:

```go
func TestBannerContainsConfiguredAddress(t *testing.T) {
    output := Banner("v0.1.0", 8888)
    require.Contains(t, output, ":8888")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./core -run TestBannerContainsConfiguredAddress -v`
Expected: FAIL because `Banner` and startup helpers do not exist.

- [ ] **Step 3: Write minimal implementation**

Implement:

- `core.Zap()` with JSON/console mode support
- `core.Banner(version string, addr int) string`
- `core.RunServer()` using `http.Server`
- `initialize.OtherInit()` to set Gin mode and invoke validator setup

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./core -run TestBannerContainsConfiguredAddress -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/core/*.go server/initialize/init.go
git commit -m "Wire startup logging and server boot flow"
```

## Task 4: Add GORM MySQL initialization

**Files:**
- Create: `server/initialize/gorm.go`
- Create: `server/initialize/gorm_mysql.go`
- Modify: `server/global/global.go`

- [ ] **Step 1: Write the failing DSN test**

Create `server/initialize/gorm_mysql_test.go`:

```go
func TestBuildMySQLDSN(t *testing.T) {
    cfg := config.Mysql{Username: "root", Password: "secret", Path: "127.0.0.1", Port: "3306", DBName: "optimus", Config: "charset=utf8mb4&parseTime=True&loc=Local"}
    require.Equal(t, "root:secret@tcp(127.0.0.1:3306)/optimus?charset=utf8mb4&parseTime=True&loc=Local", buildMySQLDSN(cfg))
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./initialize -run TestBuildMySQLDSN -v`
Expected: FAIL because the MySQL initialization code does not exist.

- [ ] **Step 3: Write minimal implementation**

Implement:

- `initialize.Gorm()` dispatcher guarded by configured db type
- `initialize.GormMysql()` opening a `gorm.DB`
- DSN builder helper
- global DB assignment

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./initialize -run TestBuildMySQLDSN -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/initialize/gorm*.go server/initialize/gorm_mysql_test.go server/global/global.go
git commit -m "Add MySQL-backed GORM initialization"
```

## Task 5: Add Redis and Casbin initialization

**Files:**
- Create: `server/initialize/redis.go`
- Create: `server/initialize/casbin.go`
- Modify: `server/global/global.go`

- [ ] **Step 1: Write the failing initialization guard tests**

Create `server/initialize/deps_test.go`:

```go
func TestRedisInitSkipsWhenDisabled(t *testing.T) {
    global.Config.System.UseRedis = false
    require.NotPanics(t, func() { Redis() })
}

func TestCasbinInitSkipsWhenDisabled(t *testing.T) {
    global.Config.System.UseCasbin = false
    require.NotPanics(t, func() { Casbin() })
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./initialize -run 'TestRedisInitSkipsWhenDisabled|TestCasbinInitSkipsWhenDisabled' -v`
Expected: FAIL because the initialization functions are missing.

- [ ] **Step 3: Write minimal implementation**

Implement:

- `initialize.Redis()` with `go-redis/v9`
- `initialize.Casbin()` with file-model bootstrap
- disabled-mode guards based on config flags

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./initialize -run 'TestRedisInitSkipsWhenDisabled|TestCasbinInitSkipsWhenDisabled' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/initialize/redis.go server/initialize/casbin.go server/initialize/deps_test.go server/global/global.go
git commit -m "Wire optional Redis and Casbin dependencies"
```

## Task 6: Implement the route registration skeleton

**Files:**
- Create: `server/router/enter.go`
- Create: `server/router/system/enter.go`
- Create: `server/initialize/router.go`
- Create: `server/middleware/logger.go`
- Create: `server/middleware/recovery.go`
- Create: `server/middleware/jwt.go`
- Create: `server/middleware/casbin_rbac.go`
- Test: `server/router/router_test.go`

- [ ] **Step 1: Write the failing route registration test**

Create `server/router/router_test.go`:

```go
func TestRoutersExposeHealthAndPing(t *testing.T) {
    engine := initialize.Routers()

    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
    engine.ServeHTTP(w, req)
    require.Equal(t, http.StatusOK, w.Code)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./router -run TestRoutersExposeHealthAndPing -v`
Expected: FAIL because router setup does not exist yet.

- [ ] **Step 3: Write minimal implementation**

Implement:

- route group aggregator in `router/enter.go`
- public/private groups in `initialize.Routers()`
- logger and recovery middleware
- JWT/Casbin placeholder middleware that currently passes through when not configured

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./router -run TestRoutersExposeHealthAndPing -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/router/*.go server/router/system/*.go server/initialize/router.go server/middleware/*.go server/router/router_test.go
git commit -m "Register the first Gin router skeleton"
```

## Task 7: Add the first system API and response helpers

**Files:**
- Create: `server/api/v1/system/health.go`
- Create: `server/api/v1/system/base.go`
- Create: `server/service/enter.go`
- Create: `server/service/system/health.go`
- Create: `server/model/common/response.go`
- Create: `server/utils/response.go`
- Create: `server/utils/claims.go`
- Modify: `server/router/system/enter.go`

- [ ] **Step 1: Write the failing handler behavior test**

Extend `server/router/router_test.go`:

```go
func TestPingRouteReturnsExpectedPayload(t *testing.T) {
    engine := initialize.Routers()

    w := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/api/v1/base/ping", nil)
    engine.ServeHTTP(w, req)
    require.Equal(t, http.StatusOK, w.Code)
    require.Contains(t, w.Body.String(), "\"message\":\"pong\"")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./router -run TestPingRouteReturnsExpectedPayload -v`
Expected: FAIL because no handlers are attached yet.

- [ ] **Step 3: Write minimal implementation**

Implement:

- system service returning health and ping payloads
- API handlers using a shared response helper
- route registration for `/health` and `/base/ping`

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./router -run 'TestRoutersExposeHealthAndPing|TestPingRouteReturnsExpectedPayload' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/api/v1/system/*.go server/service/*.go server/service/system/*.go server/model/common/response.go server/utils/*.go server/router/system/enter.go server/router/router_test.go
git commit -m "Expose health and ping system endpoints"
```

## Task 8: Verify the full backend skeleton

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Run the backend unit test suite**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 2: Run a build verification**

Run: `go build ./...`
Expected: PASS

- [ ] **Step 3: Smoke-test startup**

Run: `go run .`
Expected: server starts, prints configured address banner, and serves `/api/v1/health`

- [ ] **Step 4: Document the startup flow**

Update `README.md` with:

- backend directory layout
- config file path
- start command
- health endpoint

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "Document and verify the backend skeleton"
```

## Review Checklist

- Confirm all file paths match the approved spec
- Confirm the plan preserves `server/main.go`, `server/core`, `server/router`, and `gin-vue-admin`-style `initialize` and `global`
- Confirm Redis and Casbin are optional at startup
- Confirm no auth business module work sneaks into the first iteration
- Confirm each test command is narrow enough for TDD and each task ends with a concrete verification step
