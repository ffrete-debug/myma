# Base image pinning
# ------------------
# Every stage is pinned to an exact patch-level tag *and* to the digest of that
# tag's multi-arch manifest index, so a rebuild of a given commit resolves to
# byte-identical base images. `alpine:latest` previously meant the runtime OS -
# and with it the Node.js major version installed by apk - could change under us
# between two builds of the same source.
#
# The digests are manifest *indexes*, not per-architecture manifests, so the
# multi-platform buildx build in .github/workflows/docker-build.yml still works.
# Refresh them together with the tag:
#
#   docker buildx imagetools inspect alpine:3.22 --format '{{.Manifest.Digest}}'
#
# golang 1.24.4 matches the `toolchain go1.24.4` directive in server/go.mod, so
# the build never downloads a second Go toolchain.
# node 20 is kept as-is from before this change. Worth knowing: CI runs the UI
# job on Node 22, and the runtime stage below serves the standalone output with
# whatever Node alpine 3.22 ships (22.x). Aligning all three on one major is a
# sensible follow-up but is a behaviour change, so it is not done here.

# Stage 1: Build Go binary
FROM golang:1.24.4-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder-go
WORKDIR /app/server

# go.mod/go.sum are copied on their own so the module download lands in its own
# layer that is only invalidated when the dependency set changes. That split is
# pointless without an explicit `go mod download` between the two COPYs: the
# `COPY server/ .` below busts the cache on every source edit, and the modules
# would then be re-fetched as a side effect of `go build`.
COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .
RUN go build -o /app/bin/ark-commander .

# Stage 2: Build Next.js frontend
FROM node:20.19.0-alpine@sha256:8bda036ddd59ea51a23bc1a1035d3b5c614e72c01366d989f4120e8adca196d4 AS builder-ui
WORKDIR /app/ui

# `npm ci`, not `npm install`. npm ci installs exactly what package-lock.json
# pins and fails loudly when the lockfile and package.json disagree, whereas
# npm install is free to resolve newer versions and rewrite the lockfile. With
# npm install here the shipped image could contain different dependency
# versions than the tree CI tested - CI runs npm ci.
COPY ui/package.json ui/package-lock.json ./
RUN npm ci

COPY ui/ ./

# NEXT_PUBLIC_* is a BUILD-time concept, not a runtime one. During compilation
# Next.js replaces every `process.env.NEXT_PUBLIC_FOO` reference with a string
# literal, in the client bundle *and* in the server bundle that ends up in
# .next/standalone. Setting NEXT_PUBLIC_API_BASE with ENV in the runtime stage,
# or through the compose `env_file`, therefore cannot work: by then the value
# is already baked into the compiled output and the variable is simply never
# read. It has to be supplied here, before `npm run build`:
#
#   docker build --build-arg NEXT_PUBLIC_API_BASE=http://ark.example:8080/api .
#
# or from docker-compose.yml via `build.args`. The default below reproduces the
# value the runtime ENV used to carry, so a plain build behaves as before.
ARG NEXT_PUBLIC_API_BASE=http://localhost:8080/api
ENV NEXT_PUBLIC_API_BASE=${NEXT_PUBLIC_API_BASE}

RUN npm run build

# Stage 3: Final production image
FROM alpine:3.24@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
WORKDIR /app

# nodejs serves the Next.js standalone output. tini is a real init: it runs as
# PID 1, reaps orphaned processes and forwards signals to entrypoint.sh (see
# ENTRYPOINT). busybox already supplies the wget used by the compose
# healthcheck, so nothing else is needed.
# su-exec drops privileges after the entrypoint has fixed ownership on the
# bind-mounted volumes, which it can only do as root.
RUN apk add --no-cache nodejs tini su-exec

# The app drives the host Docker daemon through /var/run/docker.sock, which is
# group-owned by the host's "docker" group. That GID differs per host, so the
# runtime user must join a group with the *host's* GID or it cannot open the
# socket. Override at build time, e.g.:
#   docker build --build-arg DOCKER_GID=$(stat -c '%g' /var/run/docker.sock) .
# docker-compose.yml passes this through from the DOCKER_GID env var.
ARG DOCKER_GID=999
RUN set -eux; \
    if ! getent group "${DOCKER_GID}" >/dev/null; then addgroup -g "${DOCKER_GID}" docker; fi; \
    adduser -D -u 10001 arkcommander; \
    addgroup arkcommander "$(getent group "${DOCKER_GID}" | cut -d: -f1)"

COPY --from=builder-go /app/bin/ark-commander /app/bin/ark-commander
COPY --from=builder-ui /app/ui/.next/standalone ./
COPY --from=builder-ui /app/ui/.next/static ./.next/static
COPY --from=builder-ui /app/ui/public ./public
COPY --from=builder-ui /app/ui/messages ./messages
COPY entrypoint.sh /app/entrypoint.sh

# /data holds the SQLite DB; the app also creates works/ and backups/ under the
# working directory at runtime, so both must be owned by the runtime user.
# chmod here rather than `COPY --chmod` so the build does not require BuildKit.
RUN mkdir -p /data \
    && chmod 0755 /app/entrypoint.sh \
    && chown -R arkcommander:arkcommander /data /app

EXPOSE 8080 3000
ENV DB_PATH=/data/ark-commander.db
# No leading colon: server/config/config.go assigns SERVER_PORT to ServerPort
# verbatim.
ENV SERVER_PORT=8080
# Docker sets HOSTNAME to the container ID, and the Next.js standalone server
# binds to whatever HOSTNAME says. Pin it so the UI listens on all interfaces.
ENV HOSTNAME=0.0.0.0
# NEXT_PUBLIC_API_BASE is intentionally NOT set here - see stage 2. A runtime
# ENV for a NEXT_PUBLIC_* variable is a silent no-op; use --build-arg.

# Deliberately NOT `USER arkcommander`.
#
# The entrypoint starts as root, chowns the bind-mounted /data and /app/backups
# to the app user, and only then drops to arkcommander via su-exec. Setting USER
# here instead looks safer but is broken: a bind mount replaces whatever the
# image had at that path, so the build-time chown above is masked by the host
# directory's ownership (root:root, as Docker creates it). The app then cannot
# create its SQLite database and crash-loops with
#   unable to open database file: out of memory (14)
# The process still ends up running as arkcommander - see entrypoint.sh.

# tini is PID 1 so signals are forwarded and zombies are reaped; entrypoint.sh
# supervises the two long-lived processes. The previous
# `sh -c "... & ... & wait"` did neither: sh as PID 1 swallowed SIGTERM, so
# compose's stop_grace_period always expired into SIGKILL - killing SQLite
# mid-write and bypassing the Go server's srv.Shutdown - and `wait` kept the
# container "up" on a dead backend as long as the Node process survived.
ENTRYPOINT ["/sbin/tini", "--", "/app/entrypoint.sh"]
