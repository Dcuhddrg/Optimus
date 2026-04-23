# Optimus

Backend bootstrap work is tracked under `server/`.

Current backend skeleton scope:
- `server/main.go` as the entrypoint
- `server/core` for runtime boot logic
- `server/router` for route registration

Backend quick start:
- Config file: `server/config.yaml`
- Start command: `cd server && go run .`
- Temporary port override: `cd server && OPTIMUS_SYSTEM_ADDR=18089 go run .`
- Health endpoint: `GET /api/v1/health`
- Ping endpoint: `GET /api/v1/ping`
- Plan-aligned ping endpoint: `GET /api/v1/base/ping`
