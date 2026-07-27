# Architecture

Reference notes for developers working on the codebase. For a feature overview
and the REST endpoint list see [`../README.md`](../README.md); for running the
stack see [`DEPLOYMENT.md`](DEPLOYMENT.md).

- Backend: Go 1.24 + Gin + GORM + SQLite (`github.com/glebarez/sqlite`), module `ark-server-commander`
- Frontend: Next.js 15 + React 19 + TypeScript + Tailwind CSS 4
- Shipped as a single Docker image (Go API + Next.js server, see `Dockerfile` / `entrypoint.sh`)

## Repository layout

```
server/
├── main.go                    # entrypoint, swagger setup
├── config/                    # JWT_SECRET / DB_PATH / SERVER_PORT env handling
├── controllers/               # HTTP handlers: audit, auth, images, player, plugins, rcon, servers
├── routes/                    # route registration (routes.go, update_routes.go)
├── middleware/                # auth, audit, rate_limiter, request_id, security, timeout
├── models/                    # GORM models (server, user, audit_log, update_status)
├── service/
│   ├── backup/                # volume backup/restore
│   ├── docker_manager/        # Docker SDK wrappers (container, exec, volume, image, rollback)
│   ├── player/                # player listing/management
│   ├── rcon/                  # RCON client wrapper
│   ├── server/                # server CRUD, start/stop/recreate with rollback
│   └── update/                # base-image update monitoring
├── utils/                     # JWT, INI validation, password hashing, logging, error codes
├── database/                  # SQLite + GORM migrations
├── websocket/                 # WebSocket hub and per-feature handlers
└── docs/                      # Swagger docs (auto-generated; HTTP endpoints only)

ui/src/
├── app/(auth)/                # login, init
├── app/(protected)/           # servers, plugins, audit-logs, home
├── app/api/                   # Next.js proxy routes to the Go API
├── components/servers/        # ServerCard, INI editors, RCONConsole
├── components/docker/         # ImageStatus, ImageUpdateConfirmModal
├── components/audit/          # audit log viewer
├── components/ui/             # Radix UI primitives
├── stores/                    # auth store, servers store
├── hooks/, lib/, config/      # axios client, ark-settings, locale, utils
└── i18n.ts                    # i18n (en, zh, pt-BR)
```

## Auth and security model

- JWT: 24h access token + 30d refresh token; logout adds the token to a blacklist.
- `JWT_SECRET` must be at least 32 characters and is rejected if it matches a
  known weak pattern — see `server/config/config.go`.
- Passwords are hashed with bcrypt (`server/utils/password.go`).
- Audit logging covers sensitive operations (create, delete, start, stop,
  `rcon.execute`) via `server/middleware/audit.go`.
- INI content is validated before it is persisted
  (`server/utils/config_files.go`) for both `GameUserSettings.ini` and `Game.ini`.
- Port conflict detection runs before a server container is created.

## Key patterns

- **Transactional Docker rollback** — destructive operations (create, start,
  stop, recreate) record compensating actions and unwind on failure. See
  `server/service/docker_manager/container_with_rollback.go` and `rollback.go`.
- **WebSocket hub** — `server/websocket/hub.go` fans out server status updates;
  per-feature handlers live alongside it.
- **Standardized errors** — `server/utils/errors.go` plus
  `server/utils/error_codes.go` define the `APIError` shape returned by handlers.

## WebSocket endpoints

These are not covered by Swagger (which only documents HTTP routes). All three
sit behind `AuthMiddleware`, which accepts the JWT via the `Authorization`
header or a `?token=` query parameter (browsers cannot set headers on a
WebSocket handshake).

| Route | Purpose |
|---|---|
| `GET /api/ws/updates/:id` | Base-image update / server status push |
| `GET /api/ws/rcon/:id` | Live RCON session (terminal I/O) |
| `GET /api/ws/logs/:id` | Container log tailing |

## RCON

Two entry points share `service/rcon`:

- `POST /api/servers/:id/rcon/execute` — one-shot command, returns
  `{"message": "...", "data": {"output": "..."}}`. Intended for occasional
  ad-hoc commands; the handler is deliberately thin (bind, dispatch, audit).
- `GET /api/ws/rcon/:id` — interactive session backing the xterm.js console in
  `ui/src/components/servers/RCONConsole.tsx`.

Session limits enforced on the WebSocket path (`server/websocket/rcon_handler.go`):

| Limit | Value |
|---|---|
| Max inbound command size | 1 KB |
| Idle read deadline | 30s |
| Write deadline | 10s |
| Minimum interval between commands | 250ms |
| Max output per response | 64 KB |

Every command is written to the audit log as `rcon.execute` with the server ID,
user ID and client IP.
