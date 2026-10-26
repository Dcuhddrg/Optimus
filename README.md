# Optimus

项目方向、当前能力、架构约定与问题反馈统一从 [项目文档入口](docs/README.md) 阅读。以下保留启动说明；早期骨架范围不代表当前能力全貌，实际状态见 [能力状态表](docs/project-status.md)。

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

Frontend quick start:
- Frontend directory: `web/`
- Package manager: pnpm（遵循项目架构约定）
- Install command: `cd web && pnpm install`
- Dev command（需要本地热更新时）: `cd web && pnpm dev`
- 默认运行效果验证和 Mac 访问按 [架构部署约定](docs/architecture/current-architecture.md) 使用 Docker Compose。
