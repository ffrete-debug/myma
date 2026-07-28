#!/bin/sh
#
# Container supervisor for the ARK Commander image.
#
# The image runs two long-lived processes:
#   * the Go API            -> /app/bin/ark-commander
#   * the Next.js frontend  -> node /app/server.js  (standalone output)
#
# Docker only ever signals PID 1, so something has to
#   (a) forward SIGTERM/SIGINT to *both* children, and
#   (b) exit as soon as *either* child dies.
#
# (a) matters because the Go server installs a graceful-shutdown handler
# (srv.Shutdown) and closes the SQLite database on the way out. Without signal
# propagation, docker-compose's stop_grace_period simply expires into SIGKILL
# and the database is killed mid-write. (b) matters because otherwise a crashed
# Go API leaves a container that still reports "up" and still passes nothing
# but the Node process, with `restart: unless-stopped` never kicking in.
#
# Zombie reaping is NOT this script's job. The Dockerfile runs it under tini
# (ENTRYPOINT ["/sbin/tini", "--", "/app/entrypoint.sh"]); tini is PID 1 and
# reaps orphans.

set -eu

# Defaults mirror the Dockerfile ENVs so the script is also usable standalone.
# SERVER_PORT has no leading colon: server/config/config.go assigns it to
# ServerPort verbatim.
: "${DB_PATH:=/data/ark-commander.db}"
: "${SERVER_PORT:=8080}"
: "${PORT:=3000}"
# Only a fallback: Docker presets HOSTNAME to the container ID, so in the image
# the authoritative value is `ENV HOSTNAME=0.0.0.0` in the Dockerfile. Without
# it the Next.js standalone server binds to the container ID's address only.
: "${HOSTNAME:=0.0.0.0}"
export DB_PATH SERVER_PORT PORT HOSTNAME

# --- privilege handling -----------------------------------------------------
# Start as root so the bind-mounted volumes can be made writable by the app
# user, then immediately drop privileges.
#
# This has to happen at RUNTIME on every start: a bind mount carries the HOST
# directory's ownership (root:root, as Docker creates it), which masks any chown
# done at build time. Relying on a build-time chown plus `USER` in the Dockerfile
# left the app unable to create its SQLite database, and it crash-looped with
#   unable to open database file: out of memory (14)
APP_USER="${APP_USER:-arkcommander}"

if [ "$(id -u)" = "0" ]; then
    for dir in "$(dirname "$DB_PATH")" "${BACKUP_DIR:-/app/backups}"; do
        mkdir -p "$dir" 2>/dev/null || true
        if ! chown -R "$APP_USER:$APP_USER" "$dir" 2>/dev/null; then
            echo "[entrypoint] warning: could not chown $dir; the app may not be able to write there"
        fi
    done

    echo "[entrypoint] dropping privileges to $APP_USER"
    exec su-exec "$APP_USER" "$0" "$@"
fi
# --- running as the app user from here ----------------------------------------


# How long to let the children finish after SIGTERM. Keep this below the
# stop_grace_period in docker-compose.yml (10s) so the SIGKILL fallback below
# is ours and not Docker's.
: "${SHUTDOWN_GRACE_SECONDS:=8}"

api_pid=''
ui_pid=''
stopping=0
exit_code=0

log() {
    echo "[entrypoint] $*" >&2
}

# `kill` is guarded everywhere: under `set -e` a bare failing `kill -0` on an
# already-dead pid would abort the script.
alive() {
    [ -n "$1" ] && kill -0 "$1" 2>/dev/null
}

signal_children() {
    if alive "$api_pid"; then kill -"$1" "$api_pid" 2>/dev/null || true; fi
    if alive "$ui_pid"; then kill -"$1" "$ui_pid" 2>/dev/null || true; fi
}

on_term() {
    if [ "$stopping" -eq 0 ]; then
        stopping=1
        log "received termination signal, stopping children"
        signal_children TERM
    fi
}

trap on_term TERM INT

# `wait -n` is not POSIX and busybox ash does not reliably provide it, so the
# loops below poll. The sleep is backgrounded and `wait`ed on rather than run in
# the foreground: `wait` is interruptible, so an incoming signal runs the trap
# immediately instead of after the current tick.
nap() {
    sleep "$1" &
    wait "$!" 2>/dev/null || true
}

log "starting Go API (SERVER_PORT=${SERVER_PORT}, DB_PATH=${DB_PATH})"
/app/bin/ark-commander &
api_pid=$!

log "starting Next.js server (HOSTNAME=${HOSTNAME}, PORT=${PORT})"
node /app/server.js &
ui_pid=$!

# Run until a signal arrives or one of the children exits.
while [ "$stopping" -eq 0 ]; do
    if ! alive "$api_pid"; then
        log "Go API (pid ${api_pid}) exited; shutting down the container"
        exit_code=1
        break
    fi
    if ! alive "$ui_pid"; then
        log "Next.js server (pid ${ui_pid}) exited; shutting down the container"
        exit_code=1
        break
    fi
    nap 1
done

# Whether we got here from a signal or from a dead child, both children must go.
signal_children TERM

waited=0
while [ "$waited" -lt "$SHUTDOWN_GRACE_SECONDS" ]; do
    if ! alive "$api_pid" && ! alive "$ui_pid"; then
        break
    fi
    nap 1
    waited=$((waited + 1))
done

if alive "$api_pid" || alive "$ui_pid"; then
    log "children still running after ${SHUTDOWN_GRACE_SECONDS}s, sending SIGKILL"
    signal_children KILL
fi

# Reap whatever is left so tini does not have to; failures are expected for a
# child the shell has already collected.
wait "$api_pid" 2>/dev/null || true
wait "$ui_pid" 2>/dev/null || true

log "exiting with status ${exit_code}"
exit "$exit_code"
