# Security Policy

## Reporting a Vulnerability

Report vulnerabilities to the repository maintainer via GitHub Issues (private).
Do not post public issues for security-critical bugs.

## Best Practices

- Use a strong JWT_SECRET (min 32 chars, generated via `openssl rand -base64 48`)
- Keep Docker and host system updated
- Restrict CORS origins in production
- Use HTTPS in production
