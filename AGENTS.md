# AGENTS.md

## Project: myma (ark-commander)

ARK Server Commander — web UI for managing ARK: Survival Evolved servers in Docker.
- Backend: Go 1.24 + Gin + GORM + SQLite (`github.com/glebarez/sqlite`)
- Frontend: Next.js 15 + React 19 + TypeScript + Tailwind CSS 4
- Go module: `ark-server-commander`
- Docker compose: `docker-compose up -d` (API :8080, Frontend :3000)

## Developer Commands

```bash
# Backend (Go)
cd server && go run main.go                          # dev
cd server && go build -o ../bin/ark-commander .     # build
cd server && go test ./... -v                        # test all
cd server && go vet ./...                            # lint
cd server && go fmt ./... && go vet ./...            # fmt + lint check

# Frontend (Next.js)
cd ui && npm install                                 # deps
cd ui && npm run dev                                 # dev server
cd ui && npm run build                               # production build
cd ui && npm run lint                                # lint

# Docker
docker-compose up -d                                 # start all
docker-compose down                                  # stop all
```

## Architecture

```
server/
├── main.go                    # entrypoint, swagger setup
├── config/config.go           # JWT_SECRET env (min 32 chars), DB_PATH, SERVER_PORT
├── controllers/               # HTTP handlers (auth, servers, images, plugins, rcon)
├── routes/                    # route registration (routes.go, update_routes.go)
├── middleware/                # auth, audit, rate_limit, request_id, security, timeout
├── models/                    # GORM models (server, user, audit_log, update_status)
├── service/
│   ├── docker_manager/        # Docker SDK wrappers (container, exec, volume, image, rollback)
│   ├── server/                # server CRUD, start/stop/recreate logic with rollback
│   └── update/                # Docker image update monitoring
├── utils/                     # JWT, INI validation, password hashing, logging, error codes
├── database/                  # SQLite + GORM migrations
├── websocket/                 # WebSocket hub for real-time updates
└── docs/                      # Swagger docs (auto-generated)

ui/
├── src/app/(auth)/            # login, init pages
├── src/app/(protected)/       # servers, plugins, home
├── src/components/servers/    # ServerCard, editors, RCONConsole
├── src/components/docker/     # ImageStatus, ImageUpdateConfirmModal
├── src/components/ui/         # Radix UI primitives (button, card, dialog, etc.)
├── src/i18n.ts                # i18n (en, zh, pt-BR)
├── src/stores/                # auth store, servers store
└── src/lib/                   # axios, ark-settings, locale, utils

docs/superpowers/specs/        # brainstorming design docs
docs/superpowers/plans/        # implementation plans
```

## Auth & Security

- JWT tokens: 24h access + 30d refresh, blacklist on logout
- Passwords: bcrypt
- Audit logging: all sensitive operations (create, delete, start, stop, rcon_command)
- INI validation before save (GameUserSettings.ini, Game.ini)
- Port conflict detection for servers
- `JWT_SECRET` must be ≥32 chars, no weak patterns allowed (see `server/config/config.go`)

## Key Patterns

- **RCON**: existing infrastructure (port config, Docker mapping, info endpoint at `GET /servers/:id/rcon`). Live command execution still in progress.
- **Docker rollback**: all destructive operations (create, start, stop, recreate) use transactional rollback via `container_with_rollback.go` and `rollback.go`.
- **WebSocket hub**: `server/websocket/hub.go` — used for server status updates; extend for RCON streams.
- **INI parsing**: `server/utils/config_files.go` — validates INI format before persisting.
- **Error handling**: standardized `APIError` in `server/utils/errors.go` + `server/utils/error_codes.go`.

## RCON Integration (in progress)

- Goal: live RCON console embedded in server detail page (collapsible panel)
- Backend: RCON client wrapper + HTTP `/servers/:id/rcon/execute` + WS `/ws/rcon/:serverID`
- Frontend: xterm.js terminal component (`RCONConsole.tsx`)
- Auth: reuses JWT session
- Audit: all RCON commands logged

## Skills & Methodology

- Superpowers opencode plugin: `superpowers@git+https://github.com/obra/superpowers.git`
- Skills: brainstorming, writing-plans, subagent-driven-development, test-driven-development, systematic-debugging, requesting-code-review, executing-plans
- Frontend Design skill: `skills/anthropics-skills/skills/frontend-design/SKILL.md`
- Always invoke brainstorming before any new feature work

## CI/CD

- GitHub Actions: `cd server → go build/test/lint`, `cd ui → npm ci/lint/build`
- Branch protection: `main` protected; use `dev/` branches for features
- PRs target `main` (not `dev`) via dev branch workflow

## Memory

- Project history tracked in `MEMORY.md` — update after every significant session