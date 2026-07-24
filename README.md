# ARK Server Commander

**Gerenciador web completo para servidores ARK: Survival Evolved em Docker**

![Go](https://img.shields.io/badge/Go-1.24-blue) ![Next.js](https://img.shields.io/badge/Next.js-15-black) ![License](https://img.shields.io/badge/license-MIT-green) [![CI](https://github.com/ffrete-debug/ark-commander/actions/workflows/ci.yml/badge.svg)](https://github.com/ffrete-debug/ark-commander/actions/workflows/ci.yml)

---

## Funcionalidades

### Gerenciamento de Servidores
- **CRUD completo** — Criar, listar, editar e excluir servidores ARK
- **Controle de ciclo de vida** — Iniciar, parar, reiniciar e reconstruir containers
- **Configuração INI** — Editor de `GameUserSettings.ini` e `Game.ini` com validação de formato
- **Múltiplos mapas** — Suporte a todos os mapas oficiais e personalizados
- **Mods** — Gerenciamento de IDs de mods Steam Workshop

### Infraestrutura Docker
- **Gerenciamento de containers** — Criação, remoção e reconstrução com rollback automático
- **Volumes persistentes** — Dados e plugins em volumes Docker separados
- **Imagens** — Pull e update da imagem base `tbro98/ase-server:latest`
- **Rollback transacional** — Reversão automática em caso de falha

### Autenticação e Segurança
- **JWT** — Tokens de acesso (24h) + refresh tokens (30d)
- **Blacklist de tokens** — Logout invalida tokens ativos
- **bcrypt** — Senhas com hash bcrypt
- **Audit logging** — Registro de todas as operações sensíveis
- **INI validation** — Validação de formato de configuração antes de salvar

### Monitoramento
- **Logs em tempo real** — Acesso aos logs do container via API
- **Status de servidor** — Acompanhamento do estado (running/stopped/starting)
- **WebSocket** — Notificações push de status de atualização
- **RCON** — Informações de conexão remota

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.24 + Gin + GORM |
| Banco | SQLite (via glebarez/sqlite) |
| Frontend | Next.js 15 + React 19 + TypeScript |
| Estilo | Tailwind CSS 4 |
| Infra | Docker + Docker Compose |
| Auth | JWT (golang-jwt/v5) + bcrypt |
| Logs | Zap (estruturado) |
| API Docs | Swagger/OpenAPI |

## Quick Start

```bash
# 1. Clone
git clone https://github.com/ffrete-debug/ark-commander.git
cd ark-commander

# 2. Configure
cp .env.example .env
# Edite JWT_SECRET (mínimo 32 caracteres)

# 3. Execute
docker-compose up -d

# 4. Acesse
# http://localhost:8080 (API + Swagger)
# http://localhost:3000 (Frontend)
```

### Desenvolvimento

```bash
# Backend
cd server
export JWT_SECRET='sua-chave-secreta-aqui-com-pelo-menos-32-caracteres'
go run main.go

# Frontend
cd ui
npm install
npm run dev
```

## API

Endpoints: `http://localhost:8080/api`

### Autenticação
| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/auth/check-init` | Verifica se sistema foi inicializado |
| POST | `/auth/init` | Cria admin inicial |
| POST | `/auth/login` | Login |
| POST | `/auth/refresh` | Renova token |
| POST | `/auth/logout` | Logout (invalida token) |

### Servidores (requer auth)
| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/servers` | Lista servidores |
| POST | `/servers` | Cria servidor |
| GET | `/servers/:id` | Detalhes |
| PUT | `/servers/:id` | Atualiza |
| DELETE | `/servers/:id` | Remove |
| POST | `/servers/:id/start` | Inicia |
| POST | `/servers/:id/stop` | Para |
| POST | `/servers/:id/restart` | Reinicia |
| POST | `/servers/:id/recreate` | Reconstrói container |
| GET | `/servers/:id/rcon` | Info RCON |
| GET | `/servers/:id/logs` | Logs do container |

### Imagens (requer auth)
| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/images/status` | Status da imagem |
| POST | `/images/pull` | Pull manual |
| GET | `/images/check-updates` | Verifica atualizações |
| POST | `/images/update` | Atualiza imagem |
| GET | `/images/affected` | Servidores afetados |

### Plugins (requer auth)
CRUD completo de arquivos de plugins via API REST.

## Estrutura

```
├── server/               # Backend Go
│   ├── config/           # Configuração (JWT, env)
│   ├── controllers/      # Handlers HTTP
│   ├── database/         # SQLite + migrations
│   ├── middleware/        # Auth, audit
│   ├── models/           # GORM models
│   ├── routes/           # Registro de rotas
│   ├── service/          # Lógica de negócio
│   │   ├── docker_manager/  # Docker SDK
│   │   ├── server/          # CRUD servidores
│   │   └── update/          # Monitor de updates
│   ├── utils/            # Helpers (INI, JWT, log, erros)
│   └── websocket/        # WebSocket hub
├── ui/                   # Frontend Next.js
│   └── src/
│       ├── app/          # Páginas (auth + protected)
│       └── components/   # Componentes React
├── docker-compose.yml
├── Dockerfile
└── .github/workflows/    # CI
```

## Changelog

Veja [CHANGELOG.md](CHANGELOG.md) para histórico completo de versões.

## Licença

MIT

## Roadmap

- [ ] RCON integration for live server commands
- [ ] Automated backups to cloud storage
- [ ] Mod management UI (Steam Workshop browser)
- [ ] Multi-language support (en, zh, pt-BR)
- [ ] Server metrics dashboard (CPU, RAM, players)
- [ ] Kubernetes support

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Security

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

- [ ] RCON integration for live server commands
- [ ] Automated backups to cloud storage
- [ ] Mod management UI (Steam Workshop browser)
- [ ] Multi-language support (en, zh, pt-BR)
- [ ] Server metrics dashboard (CPU, RAM, players)
- [ ] Kubernetes support
