# ARK Server Commander

**A complete web manager for ARK: Survival Evolved servers running in Docker**

![Go](https://img.shields.io/badge/Go-1.24-blue) ![Next.js](https://img.shields.io/badge/Next.js-15-black) ![Node](https://img.shields.io/badge/Node-20%2B-green) ![License](https://img.shields.io/badge/license-MIT-green) [![CI](https://github.com/ffrete-debug/ark-commander/actions/workflows/ci.yml/badge.svg)](https://github.com/ffrete-debug/ark-commander/actions/workflows/ci.yml)

**English** | [Português (pt-BR)](README-pt-BR.md) | [中文](README-zh.md)

---

## Features

### Server management
- **Full CRUD** — create, list, edit and delete ARK servers
- **Lifecycle control** — start, stop, restart and rebuild containers
- **INI configuration** — `GameUserSettings.ini` and `Game.ini` editor with format validation
- **Multiple maps** — every official map plus custom ones
- **Mods** — manage Steam Workshop mod IDs

### Docker infrastructure
- **Container management** — create, remove and rebuild with automatic rollback
- **Persistent volumes** — game data and plugins in separate Docker volumes
- **Images** — pull and update the `tbro98/ase-server:latest` base image
- **Transactional rollback** — automatic reversal on failure

### Authentication and security
- **JWT** — access tokens (24h) + refresh tokens (30d)
- **Token blacklist** — logout invalidates live tokens
- **bcrypt** — password hashing
- **Audit logging** — every sensitive operation is recorded
- **INI validation** — configuration format is validated before it is saved

### Monitoring
- **Live logs** — container logs exposed over the API
- **Server status** — running / stopped / starting tracking
- **WebSocket** — push notifications for update status
- **RCON** — remote connection details

### Internationalisation
- **UI available in English, 中文 and Português (pt-BR)** — selected per user and
  stored in the `NEXT_LOCALE` cookie; the default is English

## Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.24 + Gin + GORM |
| Database | SQLite (via glebarez/sqlite) |
| Frontend | Next.js 15 + React 19 + TypeScript |
| Styling | Tailwind CSS 4 |
| i18n | next-intl (en, zh, pt-BR) |
| Infra | Docker + Docker Compose |
| Auth | JWT (golang-jwt/v5) + bcrypt |
| Logging | Zap (structured) |
| API docs | Swagger/OpenAPI |

### Build toolchain

The image is a three-stage build, in this order:

**Stage 1 — Go (`golang:1.24.4-alpine`)** builds the API binary
→ **Stage 2 — Node (`node:20.19.0-alpine`)** builds the Next.js standalone bundle
→ **Stage 3 — `alpine:3.22`** runtime, which installs `nodejs` and `tini` and
runs both processes as the non-root user `arkcommander`.

| Toolchain | Version | Defined in |
|-----------|---------|-----------|
| Go (image build) | 1.24.4 | `Dockerfile` stage 1 |
| Go (CI) | 1.24 | `.github/workflows/ci.yml` |
| Node (image build) | 20.19.0 | `Dockerfile` stage 2 |
| Node (CI) | 22 | `.github/workflows/ci.yml` |
| Node (runtime) | 22.x, from Alpine 3.22's `nodejs` package | `Dockerfile` stage 3 |

Every base image is pinned to an exact patch tag so the runtime OS — and with it
the Node major version `apk` installs — cannot drift between two builds of the
same commit.

The frontend is *compiled* with Node 20 and *served* on Node 22, and CI checks
it on Node 22. **Node 20 or newer** is therefore the supported floor for local
development. (Aligning all three on one major is a worthwhile follow-up; it is a
behaviour change and deliberately out of scope here.)

## Quick start

```bash
# 1. Clone
git clone https://github.com/ffrete-debug/ark-commander.git
cd ark-commander

# 2. Configure
cp .env.example .env

# 3. Generate a JWT secret and put it in .env
openssl rand -base64 48
# Edit .env and set JWT_SECRET=<the value you just generated>

# 4. Run
docker compose up -d

# 5. Open
# http://localhost:3000  (web UI)
# http://localhost:8080  (API + Swagger)
```

> **The backend will not start without `JWT_SECRET`.** It must be at least 32
> characters and must not contain a weak pattern (`secret`, `password`,
> `123456`, `default`, `changeme`, `test`, …). If the container crash-loops,
> `docker compose logs` will tell you exactly which rule was violated.

You may also need to set `DOCKER_GID` in `.env` — the container runs as a
non-root user and needs the host's `docker` group GID to reach
`/var/run/docker.sock`:

```bash
echo "DOCKER_GID=$(stat -c '%g' /var/run/docker.sock)" >> .env
```

## Configuration

Every variable below is read from the environment. With Docker Compose they
come from `.env` (`env_file`); see [`.env.example`](.env.example) for the
annotated template.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | **Yes** | — | JWT signing key. Minimum 32 characters, rejected if it contains a weak pattern. The process exits at startup if it is missing or invalid. Generate with `openssl rand -base64 48`. |
| `SERVER_PORT` | No | `8080` | TCP port for the Go API. Bare number, no leading colon. |
| `DB_PATH` | No | `ark_server.db` | SQLite database file. The Dockerfile overrides this to `/data/ark-commander.db`, which docker-compose maps to `./data`. |
| `CORS_ORIGIN` | No | *(empty)* | Comma-separated allowlist of exact origins permitted to call the API cross-origin. Empty means same-origin only — no `Access-Control-Allow-Origin` is emitted. **`*` is not supported.** This is the variable [SECURITY.md](SECURITY.md) means by "restrict CORS origins in production". |
| `TRUSTED_PROXIES` | No | *(empty)* | Comma-separated IPs/CIDRs whose `X-Forwarded-For` is trusted. Empty means trust nothing, so client IPs come from the socket and cannot be spoofed. |
| `LOG_LEVEL` | No | `info` | `debug`, `info`, `warn`, `error`, `panic` or `fatal`. Unrecognised values fall back to `info`. |
| `LOG_FORMAT` | No | `json` | `json` (structured, for log collectors) or `console` (coloured, human-readable). |
| `PORT` | No | `3000` | Port the Next.js standalone server binds to. |
| `HOSTNAME` | No | `0.0.0.0` | Interface the Next.js server binds to. Pinned in the Dockerfile because Docker presets `HOSTNAME` to the container ID. |
| `SHUTDOWN_GRACE_SECONDS` | No | `8` | Seconds `entrypoint.sh` waits after SIGTERM before escalating to SIGKILL. Keep it below the compose `stop_grace_period` (10s). |

Two settings are **build-time only** — putting them in `.env` as runtime
variables will not work:

| Build arg | Default | Description |
|-----------|---------|-------------|
| `DOCKER_GID` | `999` | Host `docker` group GID the non-root runtime user joins so it can open the Docker socket. `docker-compose.yml` forwards it from the environment via `${DOCKER_GID:-999}`, so setting it in `.env` *does* reach the build. |
| `NEXT_PUBLIC_API_BASE` | `http://localhost:8080/api` | Address the Next.js route handlers use to reach the Go API. `next build` inlines `NEXT_PUBLIC_*` values into the compiled output, so a runtime `ENV` is a silent no-op. Pass it with `docker build --build-arg NEXT_PUBLIC_API_BASE=…` (or compose `build.args`). Only needed if the API and UI are split across containers. |

### CORS in practice

The bundled UI calls the API through Next.js route handlers on its own origin,
so the default (empty `CORS_ORIGIN`) is correct for the standard
docker-compose deployment. Only set it when a browser on a **different** origin
must call the API directly:

```bash
CORS_ORIGIN=https://ark.example.com,https://admin.example.com
```

Entries are matched exactly — scheme, host and port, with no trailing slash and
no wildcards. A matching origin is echoed back with
`Access-Control-Allow-Credentials: true`, so keep the list as small as possible.

## Breaking changes

If you are upgrading an existing deployment, three recent changes need action:

1. **`admin_password` is no longer returned by `GET /api/servers` or
   `GET /api/servers/:id`.** The field is stripped from the server response
   payload so RCON credentials are not handed out on every list request. Read it
   from the dedicated `GET /api/servers/:id/rcon` endpoint instead. Any client
   that read `admin_password` off the list response must be updated.

2. **`CORS_ORIGIN` no longer defaults to `*`.** The default is now empty —
   same-origin only, with no `Access-Control-Allow-Origin` header emitted at
   all — and the value is a comma-separated allowlist of exact origins rather
   than a single value or a wildcard. Cross-origin browser clients that used to
   work implicitly will now be blocked until you list their origin.

3. **`TRUSTED_PROXIES` defaults to trusting nothing.** Gin no longer honours
   `X-Forwarded-For` from any source, so audit-log IPs and rate limiting use the
   real socket address. If you run behind a reverse proxy, set
   `TRUSTED_PROXIES` to that proxy's IP or CIDR, otherwise every request will
   appear to come from the proxy.

Related hardening in the same series: the container now runs as the non-root
user `arkcommander` (hence `DOCKER_GID`), and `privileged: true` has been
removed from `docker-compose.yml`.

## Development

```bash
# Backend
cd server
export JWT_SECRET="$(openssl rand -base64 48)"
export LOG_FORMAT=console     # readable logs while developing
go run main.go

# Frontend
cd ui
npm install
npm run dev
```

Handy `make` targets: `make build`, `make test`, `make lint`, `make fmt`.

## API

Base URL: `http://localhost:8080/api`

### Authentication
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/auth/check-init` | Check whether the system has been initialised |
| POST | `/auth/init` | Create the initial admin |
| POST | `/auth/login` | Log in |
| POST | `/auth/refresh` | Refresh the token |
| POST | `/auth/logout` | Log out (invalidates the token) |

### Servers (auth required)
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/servers` | List servers |
| POST | `/servers` | Create a server |
| GET | `/servers/:id` | Server details |
| PUT | `/servers/:id` | Update |
| DELETE | `/servers/:id` | Delete |
| POST | `/servers/:id/start` | Start |
| POST | `/servers/:id/stop` | Stop |
| POST | `/servers/:id/restart` | Restart |
| POST | `/servers/:id/recreate` | Rebuild the container |
| GET | `/servers/:id/rcon` | RCON details (the only endpoint that returns `admin_password`) |
| GET | `/servers/:id/logs` | Container logs |

### Images (auth required)
| Method | Route | Description |
|--------|-------|-------------|
| GET | `/images/status` | Image status |
| POST | `/images/pull` | Manual pull |
| GET | `/images/check-updates` | Check for updates |
| POST | `/images/update` | Update the image |
| GET | `/images/affected` | Affected servers |

### Plugins (auth required)
Full CRUD over plugin files via REST.

Interactive documentation: `http://localhost:8080/swagger/index.html`

## Project layout

```
├── server/               # Go backend
│   ├── config/           # Configuration (JWT, env)
│   ├── controllers/      # HTTP handlers
│   ├── database/         # SQLite + migrations
│   ├── middleware/       # Auth, audit, rate limiting
│   ├── models/           # GORM models
│   ├── routes/           # Route registration
│   ├── service/          # Business logic
│   │   ├── docker_manager/  # Docker SDK
│   │   ├── server/          # Server CRUD
│   │   └── update/          # Update monitor
│   ├── utils/            # Helpers (INI, JWT, logging, errors)
│   └── websocket/        # WebSocket hub
├── ui/                   # Next.js frontend
│   ├── messages/         # i18n catalogues (en, zh, pt-BR)
│   └── src/
│       ├── app/          # Pages (auth + protected)
│       └── components/   # React components
├── .env.example          # Environment template
├── docker-compose.yml
├── Dockerfile
└── .github/workflows/    # CI
```

## Roadmap

- [ ] RCON integration for live server commands
- [ ] Automated backups to cloud storage
- [ ] Mod management UI (Steam Workshop browser)
- [ ] Server metrics dashboard (CPU, RAM, players)
- [ ] Kubernetes support

Multi-language support (en, zh, pt-BR) is **already shipped** — see
[Internationalisation](#internationalisation).

## Documentation

- [CHANGELOG.md](CHANGELOG.md) — version history
- [CONTRIBUTING.md](CONTRIBUTING.md) — contribution guidelines
- [SECURITY.md](SECURITY.md) — vulnerability reporting and hardening notes
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) — community standards

## Translations

- [Português (pt-BR)](README-pt-BR.md)
- [中文](README-zh.md)

## License

MIT
