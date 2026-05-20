# Docker Compose Deployment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a minimal `deploy/docker-compose/` deployment for frontend, backend, and MySQL, where nginx serves the built frontend and proxies `/api` to the backend.

**Architecture:** The deployment stays intentionally narrow: four files only, three services only, no initialization SQL, no deployment README, no env file layer. The frontend is built in a Node stage and served by nginx, the backend is built from `server/` into a runnable binary, and MySQL is used only as an internal database service while backend still exposes a host port for local debugging.

**Tech Stack:** Docker Compose, nginx, Node build stage for Vite frontend, Go multi-stage backend build, MySQL official image

---

## File Map

### Create

- `deploy/docker-compose/docker-compose.yml` — compose definition for frontend, backend, and mysql
- `deploy/docker-compose/Dockerfile.frontend` — multi-stage frontend build and nginx runtime image
- `deploy/docker-compose/Dockerfile.backend` — multi-stage Go backend build and runtime image
- `deploy/docker-compose/nginx.conf` — nginx config for static file serving and `/api` reverse proxy

### Do Not Modify

- `server/**` — backend code and config stay unchanged
- `web/**` — frontend app code stays unchanged
- `README.md` — no deployment doc updates in this task
- `.codex/**` — no Codex rule changes in this task

## Task 1: Add the frontend container build

**Files:**
- Create: `deploy/docker-compose/Dockerfile.frontend`

- [ ] **Step 1: Write the failing frontend deployment checklist**

Create a temporary checklist in your notes:

```text
- Dockerfile.frontend exists
- uses Node build stage
- builds from /web
- final runtime uses nginx
- final image serves built static assets only
```

- [ ] **Step 2: Verify the frontend Dockerfile does not exist yet**

Run: `test -f deploy/docker-compose/Dockerfile.frontend && echo exists || echo missing`
Expected: `missing`

- [ ] **Step 3: Write the minimal frontend Dockerfile**

Create a multi-stage Dockerfile that:

- uses a Node image in the build stage
- copies `web/package*.json`
- installs frontend dependencies
- copies the `web/` source
- runs the frontend build
- uses an nginx image in the final stage
- copies the built frontend artifacts into the nginx web root
- copies `deploy/docker-compose/nginx.conf` into nginx config path

Do not add extra shell scripts or helper entrypoints.

- [ ] **Step 4: Verify the Dockerfile structure**

Run: `sed -n '1,220p' deploy/docker-compose/Dockerfile.frontend`
Expected: PASS with a Node build stage and an nginx runtime stage.

- [ ] **Step 5: Commit**

```bash
git add deploy/docker-compose/Dockerfile.frontend
git commit -m "feat(deploy): add frontend container build"
```

## Task 2: Add the backend container build

**Files:**
- Create: `deploy/docker-compose/Dockerfile.backend`

- [ ] **Step 1: Write the failing backend deployment checklist**

Create a temporary checklist in your notes:

```text
- Dockerfile.backend exists
- uses Go build stage
- builds from /server
- final runtime contains only needed binary/config
- runtime starts backend directly
```

- [ ] **Step 2: Verify the backend Dockerfile does not exist yet**

Run: `test -f deploy/docker-compose/Dockerfile.backend && echo exists || echo missing`
Expected: `missing`

- [ ] **Step 3: Write the minimal backend Dockerfile**

Create a multi-stage Dockerfile that:

- uses a Go image in the build stage
- copies `server/go.mod` and `server/go.sum`
- downloads Go dependencies
- copies the `server/` source
- builds the backend binary
- uses a small runtime image
- copies the built backend binary
- copies any required config file(s) needed by the runtime
- starts the backend directly

Do not add migration scripts or startup wrappers.

- [ ] **Step 4: Verify the Dockerfile structure**

Run: `sed -n '1,220p' deploy/docker-compose/Dockerfile.backend`
Expected: PASS with a Go build stage and a minimal runtime stage.

- [ ] **Step 5: Commit**

```bash
git add deploy/docker-compose/Dockerfile.backend
git commit -m "feat(deploy): add backend container build"
```

## Task 3: Add nginx reverse proxy config

**Files:**
- Create: `deploy/docker-compose/nginx.conf`

- [ ] **Step 1: Write the failing proxy checklist**

Create a temporary checklist in your notes:

```text
- nginx serves frontend static files
- /api is proxied to backend
- frontend fallback routes resolve to index.html
- no extra unrelated server blocks
```

- [ ] **Step 2: Verify the nginx config does not exist yet**

Run: `test -f deploy/docker-compose/nginx.conf && echo exists || echo missing`
Expected: `missing`

- [ ] **Step 3: Write the nginx config**

Create `nginx.conf` that:

- listens on the frontend container port
- serves the frontend build output
- falls back SPA routes to `index.html`
- proxies `/api/` to the backend service name in compose
- forwards standard proxy headers

Do not add TLS, domain redirects, or multiple virtual hosts.

- [ ] **Step 4: Verify the nginx config content**

Run: `sed -n '1,240p' deploy/docker-compose/nginx.conf`
Expected: PASS with static serving, SPA fallback, and `/api` proxy behavior.

- [ ] **Step 5: Commit**

```bash
git add deploy/docker-compose/nginx.conf
git commit -m "feat(deploy): add nginx api proxy config"
```

## Task 4: Add the compose definition

**Files:**
- Create: `deploy/docker-compose/docker-compose.yml`

- [ ] **Step 1: Write the failing compose checklist**

Create a temporary checklist in your notes:

```text
- compose defines only frontend, backend, mysql
- frontend builds from Dockerfile.frontend
- backend builds from Dockerfile.backend
- mysql uses official image
- nginx port is exposed
- backend host port is exposed for debugging
- mysql host port is not exposed
- backend receives db env vars
- mysql has a named volume
```

- [ ] **Step 2: Verify the compose file does not exist yet**

Run: `test -f deploy/docker-compose/docker-compose.yml && echo exists || echo missing`
Expected: `missing`

- [ ] **Step 3: Write the compose file**

Create `docker-compose.yml` that:

- defines `frontend`, `backend`, and `mysql` only
- builds frontend from `Dockerfile.frontend`
- builds backend from `Dockerfile.backend`
- uses an official MySQL image
- exposes frontend/nginx port to the host
- exposes backend port to the host for local debugging
- does not expose a mysql port to the host
- injects backend database environment variables
- gives mysql a named volume
- places the services on the default compose network

Do not add init SQL mounts, deployment docs, or extra helper services.

- [ ] **Step 4: Verify the compose structure**

Run: `sed -n '1,260p' deploy/docker-compose/docker-compose.yml`
Expected: PASS with exactly the intended three services and the intended port exposure rules.

- [ ] **Step 5: Commit**

```bash
git add deploy/docker-compose/docker-compose.yml
git commit -m "feat(deploy): add docker compose deployment"
```

## Task 5: Verify the deployment files are structurally correct

**Files:**
- Modify: `deploy/docker-compose/docker-compose.yml`
- Modify: `deploy/docker-compose/Dockerfile.frontend`
- Modify: `deploy/docker-compose/Dockerfile.backend`
- Modify: `deploy/docker-compose/nginx.conf`

- [ ] **Step 1: Verify only the four deployment files exist under the target directory**

Run: `find deploy/docker-compose -maxdepth 2 -type f | sort`
Expected: only the four deployment files are present.

- [ ] **Step 2: Verify compose service count and key rules**

Run:

```bash
grep -nE '^  (frontend|backend|mysql):$' deploy/docker-compose/docker-compose.yml
grep -n '3306:' deploy/docker-compose/docker-compose.yml || true
grep -n '/api' deploy/docker-compose/nginx.conf
```

Expected:

- three service definitions only
- no mysql host port mapping
- `/api` proxy rule present

- [ ] **Step 3: Verify Dockerfiles reference the correct build roots**

Run:

```bash
sed -n '1,220p' deploy/docker-compose/Dockerfile.frontend
sed -n '1,220p' deploy/docker-compose/Dockerfile.backend
```

Expected:

- frontend Dockerfile builds from `web/`
- backend Dockerfile builds from `server/`

- [ ] **Step 4: Make any minimal corrections needed**

If any structural check fails, adjust only the affected deployment file(s).

- [ ] **Step 5: Commit**

```bash
git add deploy/docker-compose
git commit -m "feat(deploy): finalize compose deployment layout"
```

## Task 6: Optional dry-run validation

**Files:**
- Modify: `deploy/docker-compose/docker-compose.yml`
- Modify: `deploy/docker-compose/Dockerfile.frontend`
- Modify: `deploy/docker-compose/Dockerfile.backend`
- Modify: `deploy/docker-compose/nginx.conf`

- [ ] **Step 1: Check whether Docker is available**

Run: `docker compose version`
Expected: either a valid Docker Compose version or a clear “not available” blocker.

- [ ] **Step 2: If Docker is available, validate the compose file**

Run: `docker compose -f deploy/docker-compose/docker-compose.yml config`
Expected: PASS with a rendered compose config.

- [ ] **Step 3: If Docker is not available, record the blocker**

If `docker compose version` fails, do not guess. Record in the handoff that runtime validation was not possible in the current environment.

- [ ] **Step 4: Make minimal fixes only if validation reveals syntax issues**

If `docker compose config` fails because of syntax or path issues, update only the deployment files as needed.

- [ ] **Step 5: Commit**

```bash
git add deploy/docker-compose
git commit -m "chore(deploy): validate compose config"
```
