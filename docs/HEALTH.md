# Health Check

The application exposes a health endpoint at `/health` that returns HTTP 200 when the service is running.

## Docker Healthcheck

Defined in `Dockerfile` and `docker-compose.yml`:
```yaml
healthcheck:
  test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
  interval: 30s
  timeout: 5s
  retries: 3
  start_period: 30s
```

## Dependencies

- SQLite database (file-based, no external service)
- Docker socket (`/var/run/docker.sock`) — checked on startup
