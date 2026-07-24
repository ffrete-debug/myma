# Deployment

## Docker Compose (recommended)

```bash
cp .env.example .env
# Edit JWT_SECRET (min 32 chars)

docker-compose up -d
```

## Manual

```bash
# Backend
cd server
JWT_SECRET=... go run main.go

# Frontend
cd ui
npm install
npm run dev
```

## Production Checklist

- [ ] Strong `JWT_SECRET` (openssl rand -base64 48)
- [ ] `CORS_ORIGIN` set to frontend domain
- [ ] `GIN_MODE=release`
- [ ] HTTPS reverse proxy (nginx/Caddy)
- [ ] Docker socket access restricted
- [ ] Persistent volume for `/data`
