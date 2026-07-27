# Stage 1: Build Go binary
FROM golang:1.24-alpine AS builder-go
WORKDIR /app/server
COPY server/go.mod server/go.sum ./
COPY server/ .
RUN go build -o /app/bin/ark-commander .

# Stage 2: Build Next.js frontend
FROM node:20-alpine AS builder-ui
WORKDIR /app/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm install
COPY ui/ ./
RUN npm run build

# Stage 3: Final production image
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache nodejs

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
# /data holds the SQLite DB; the app also creates works/ and backups/ under the
# working directory at runtime, so both must be owned by the runtime user.
RUN mkdir -p /data && chown -R arkcommander:arkcommander /data /app
EXPOSE 8080 3000
ENV DB_PATH=/data/ark-commander.db
ENV SERVER_PORT=8080
ENV NEXT_PUBLIC_API_BASE=http://localhost:8080/api
USER arkcommander
CMD ["sh", "-c", "/app/bin/ark-commander & node /app/server.js & wait"]
