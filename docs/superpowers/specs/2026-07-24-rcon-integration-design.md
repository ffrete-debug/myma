# RCON Integration — Design Spec

## Overview

A live RCON console embedded as a collapsible panel on each server's detail page, plus a command/response API endpoint — enabling users to send commands to running ARK servers and see responses in real-time.

## Subject & Audience

- **Subject**: ARK server administration via RCON protocol
- **Audience**: Server administrators managing ARK: Survival Evolved servers in Docker
- **Single job**: Let admins send RCON commands to running servers and get instant feedback without leaving the web UI

## Architecture

```
┌─────────────────────────────────────────────┐
│  Next.js UI (xterm.js terminal)             │
│  ├── Collapsible panel on server detail     │
│  ├── WebSocket → existing WS hub            │
│  └── HTTP POST for command/response          │
├─────────────────────────────────────────────┤
│  Go Backend (Gin)                           │
│  ├── WebSocket /ws/rcon/:serverID           │
│  │   └── Extend existing WS hub             │
│  ├── POST /servers/:id/rcon/execute         │
│  │   └── New RCON controller + service fn   │
│  └── RCON library (existing Go lib)         │
├─────────────────────────────────────────────┤
│  Docker Container (ARK server)              │
│  └── RCON port exposed (default: 32330)     │
└─────────────────────────────────────────────┘
```

## Key Components

### UI Layer
- **xterm.js panel** (`@xterm/xterm`): Industry-standard terminal emulator with themes, search, copy/paste
- **RCONPanel React component**: Terminal render, keyboard input, scrollback, connection status indicator
- **Placement**: Collapsible panel on server detail page (`ui/src/app/(protected)/servers/[id]/page.tsx`)
- **Connection lifecycle**: Establish on panel open, persist for the session, close on panel collapse

### API Layer

#### POST `/servers/:id/rcon/execute`
- **Auth**: JWT Bearer token (reuses existing auth)
- **Input**: `{"command": "status"}`
- **Output**: `{"output": "...", "exit_code": 0}`
- **Auditing**: Logs command + sanitized response

#### WS `/ws/rcon/:serverID`
- **Auth**: JWT via query param or header (consistent with existing WS)
- **Channels**: `rcon:input` (user → server), `rcon:output` (server → user)
- **Heartbeat**: Ping/pong to keep connection alive
- **Reconnect**: Client reconnects on disconnection, creates new RCON session

### Service Layer
- **`server/service/rcon/rcon_client.go`**: RCON client wrapper using Go RCON library (e.g., `github.com/renovatebot/rcon` or similar)
- **`server/service/server/server_service.go`** — add `ExecuteRCONCommand(userID, serverID, command) (string, error)` method
- **`server/controllers/rcon/rcon.go`**: `ExecuteRCON` HTTP handler + WebSocket handler
- **`server/routes/rcon_routes.go`**: Route registration for RCON endpoints

### Audit Layer
- Extend `server/middleware/audit.go` to log RCON commands
- Log entry: `{action: "rcon_command", server_id, user_id, command, response_length, timestamp}`
- Sanitize: Exclude admin password from audit logs

## Design Decisions

| Decision | Rationale |
|----------|-----------|
| xterm.js | Industry standard, full terminal features, good React ecosystem support (`@xterm/xterm` + `@xterm/addon-fit`) |
| Reuse WS hub | Leverages existing infrastructure, one WebSocket connection per user session, no new connections needed |
| Audit all commands | Full accountability trail, consistent with existing security model |
| JWT reuse | No new auth layer needed, consistent with existing security model |
| Per-panel RCON session | Clean lifecycle, no dangling connections, fresh auth per session |

## Non-Goals

- No multi-server RCON (one server per session)
- No command history persistence (in-memory scrollback only)
- No RCON file transfer (SFTP/bulk upload/download)
- No Kubernetes RCON (only Docker-based servers)

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| RCON library compatibility | Use well-maintained library with active community |
| WebSocket connection leaks | Implement connection timeout and cleanup on panel close |
| RCON command injection | Validate/sanitize commands server-side (allowlist unsafe commands if needed) |
| Large output flooding | Implement output truncation (max 64KB per response) |

## Success Criteria

1. User can open RCON panel on any running server's detail page
2. Terminal renders with xterm.js, accepts input, displays output in real-time
3. Command/response endpoint works for occasional commands
4. All RCON commands are audited in the audit log
5. Connection closes cleanly when panel is collapsed
6. JWT auth is enforced on all RCON endpoints
