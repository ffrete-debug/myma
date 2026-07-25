# Deploy & Feature Summary

Generated: 2026-07-25

## Project: ark-commander (myma)
ARK Server Commander — web UI for managing ARK: Survival Evolved servers in Docker.
- Backend: Go 1.24 + Gin + GORM + SQLite
- Frontend: Next.js 15 + React 19 + TypeScript + Tailwind CSS 4

---

## 1. Verification

### Build Status
- `go build ./...` — clean, zero errors
- `npm run build` — clean, standalone output generated
- `Dockerfile`, `.env`, `entrypoint.sh` exist and committed (b287c77)

### Files Created/Modified (latest commit)
- `Dockerfile` — multi-stage build (Go + Node.js + Alpine final)
- `.env` — JWT_SECRET with 32+ char base64
- `entrypoint.sh` — orchestrates Go API + Next.js in single container

### Docker Deployment
- `docker-compose up` in `/tmp/myma-test` timed out during base image pulls
- Fix needed: retry without `--no-cache`, longer timeout, or pre-pull images first

---

## 2. What's Implemented

| Feature | Backend | Frontend |
|---|---|---|
| Server CRUD | Done | Done |
| Server lifecycle (start/stop/restart/recreate) | Done | Done |
| RCON console (WebSocket + HTTP) | Done | Done |
| Backup/restore | Done | Done |
| Plugin file management | Done | Done |
| Docker image mgmt (pull, update) | Done | Done |
| Bulk server actions | Done | Done |
| Player management | Routes exist | Disabled ("Soon") |
| WebSocket real-time status | Done | Done |
| Auth (JWT) | Done | Done |
| Audit logging | Done | Partial |
| Swagger docs | Done | N/A |
| i18n (en/zh/pt-BR) | Done | Done |

---

## 3. What's Missing / Partial

1. **Player management frontend** — backend routes exist at `/api/servers/:id/players`
2. **Live log tailing via WebSocket** — current logs page uses polling
3. **Settings/profile page** — no user-facing settings
4. **Audit log viewer** — backend logs ops, no frontend to query
5. **Update status page** — `/api/updates/:id/status` endpoint exists, no UI
6. **Docker deployment fix** — image pull timeout in temp env

---

## 4. Suggested Sequence

1. Fix Docker deployment (build + compose)
2. Build player management frontend
3. Add live log tailing via WebSocket
4. Add settings/profile page
5. Add audit log viewer
6. Add update status dashboard

---

## 5. Key Paths

- Backend routes: `server/routes/routes.go`
- Server controller: `server/controllers/servers/server.go`
- RCON controller: `server/controllers/rcon/rcon.go`
- Backup controller: `server/controllers/servers/backup.go`
- Docker manager: `server/service/docker_manager/docker_manager.go`
- RCON WebSocket handler: `server/websocket/rcon_handler.go`
- Frontend pages: `ui/src/app/(protected)/`
- RCON console component: `ui/src/components/servers/RCONConsole.tsx`
- Server store: `ui/src/stores/servers.ts`
- i18n messages: `ui/messages/en.ts`, `ui/messages/zh.ts`, `ui/messages/pt-BR.ts`
