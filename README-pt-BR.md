# ARK Server Commander

**Gerenciador web completo para servidores ARK: Survival Evolved em Docker**

![Go](https://img.shields.io/badge/Go-1.24-blue) ![Next.js](https://img.shields.io/badge/Next.js-15-black) ![Node](https://img.shields.io/badge/Node-20%2B-green) ![License](https://img.shields.io/badge/license-MIT-green) [![CI](https://github.com/ffrete-debug/ark-commander/actions/workflows/ci.yml/badge.svg)](https://github.com/ffrete-debug/ark-commander/actions/workflows/ci.yml)

[English](README.md) | **Português (pt-BR)** | [中文](README-zh.md)

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

### Internacionalização
- **Interface em English, 中文 e Português (pt-BR)** — selecionada por usuário e
  armazenada no cookie `NEXT_LOCALE`; o padrão é inglês

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.24 + Gin + GORM |
| Banco | SQLite (via glebarez/sqlite) |
| Frontend | Next.js 15 + React 19 + TypeScript |
| Estilo | Tailwind CSS 4 |
| i18n | next-intl (en, zh, pt-BR) |
| Infra | Docker + Docker Compose |
| Auth | JWT (golang-jwt/v5) + bcrypt |
| Logs | Zap (estruturado) |
| API Docs | Swagger/OpenAPI |

### Toolchain de build

A imagem tem três estágios, nesta ordem:

**Estágio 1 — Go (`golang:1.24.4-alpine`)** compila o binário da API
→ **Estágio 2 — Node (`node:20.19.0-alpine`)** compila o bundle standalone do Next.js
→ **Estágio 3 — `alpine:3.22`** runtime, que instala `nodejs` e `tini` e roda os
dois processos com o usuário não-root `arkcommander`.

| Toolchain | Versão | Definido em |
|-----------|--------|-------------|
| Go (build da imagem) | 1.24.4 | `Dockerfile` estágio 1 |
| Go (CI) | 1.24 | `.github/workflows/ci.yml` |
| Node (build da imagem) | 20.19.0 | `Dockerfile` estágio 2 |
| Node (CI) | 22 | `.github/workflows/ci.yml` |
| Node (runtime) | 22.x, do pacote `nodejs` do Alpine 3.22 | `Dockerfile` estágio 3 |

Todas as imagens base estão fixadas em uma tag de patch exata, para que o SO de
runtime — e com ele a versão major do Node instalada pelo `apk` — não mude entre
dois builds do mesmo commit.

O frontend é *compilado* com Node 20 e *servido* em Node 22, e a CI verifica em
Node 22. Portanto **Node 20 ou superior** é o piso suportado para
desenvolvimento local. (Alinhar os três em um único major é um follow-up
desejável, mas é mudança de comportamento e está fora do escopo aqui.)

## Quick Start

```bash
# 1. Clone
git clone https://github.com/ffrete-debug/ark-commander.git
cd ark-commander

# 2. Configure
cp .env.example .env

# 3. Gere um JWT secret e coloque no .env
openssl rand -base64 48
# Edite o .env e defina JWT_SECRET=<o valor gerado>

# 4. Execute
docker compose up -d

# 5. Acesse
# http://localhost:3000  (interface web)
# http://localhost:8080  (API + Swagger)
```

> **O backend não sobe sem `JWT_SECRET`.** Precisa ter no mínimo 32 caracteres e
> não pode conter um padrão fraco (`secret`, `password`, `123456`, `default`,
> `changeme`, `test`, …). Se o container entrar em crash-loop,
> `docker compose logs` mostra exatamente qual regra foi violada.

Você provavelmente também precisa definir `DOCKER_GID` no `.env` — o container
roda como usuário não-root e precisa do GID do grupo `docker` do host para
acessar `/var/run/docker.sock`:

```bash
echo "DOCKER_GID=$(stat -c '%g' /var/run/docker.sock)" >> .env
```

## Configuração

Todas as variáveis abaixo vêm do ambiente. Com Docker Compose elas são lidas do
`.env` (`env_file`); veja [`.env.example`](.env.example) para o template
comentado.

| Variável | Obrigatória | Padrão | Descrição |
|----------|-------------|--------|-----------|
| `JWT_SECRET` | **Sim** | — | Chave de assinatura JWT. Mínimo 32 caracteres, rejeitada se contiver padrão fraco. O processo encerra na inicialização se estiver ausente ou inválida. Gere com `openssl rand -base64 48`. |
| `SERVER_PORT` | Não | `8080` | Porta TCP da API Go. Número puro, sem dois-pontos. |
| `DB_PATH` | Não | `ark_server.db` | Arquivo SQLite. O Dockerfile sobrescreve para `/data/ark-commander.db`, que o docker-compose mapeia para `./data`. |
| `CORS_ORIGIN` | Não | *(vazio)* | Allowlist de origens exatas, separadas por vírgula, autorizadas a chamar a API cross-origin. Vazio = apenas same-origin, sem emitir `Access-Control-Allow-Origin`. **`*` não é mais suportado.** É esta a variável que o [SECURITY.md](SECURITY.md) chama de "restrict CORS origins in production". |
| `TRUSTED_PROXIES` | Não | *(vazio)* | IPs/CIDRs, separados por vírgula, cujo `X-Forwarded-For` é confiável. Vazio = não confia em ninguém, então o IP do cliente vem do socket e não pode ser forjado. |
| `LOG_LEVEL` | Não | `info` | `debug`, `info`, `warn`, `error`, `panic` ou `fatal`. Valores desconhecidos caem em `info`. |
| `LOG_FORMAT` | Não | `json` | `json` (estruturado, para coletores) ou `console` (colorido, legível). |
| `PORT` | Não | `3000` | Porta do servidor standalone do Next.js. |
| `HOSTNAME` | Não | `0.0.0.0` | Interface em que o Next.js escuta. Fixada no Dockerfile porque o Docker define `HOSTNAME` como o ID do container. |
| `SHUTDOWN_GRACE_SECONDS` | Não | `8` | Segundos que o `entrypoint.sh` espera após o SIGTERM antes de escalar para SIGKILL. Mantenha abaixo do `stop_grace_period` do compose (10s). |

Duas configurações são **exclusivas de build** — colocá-las no `.env` como
variáveis de runtime não funciona:

| Build arg | Padrão | Descrição |
|-----------|--------|-----------|
| `DOCKER_GID` | `999` | GID do grupo `docker` do host ao qual o usuário não-root do runtime é adicionado, para conseguir abrir o socket do Docker. O `docker-compose.yml` repassa via `${DOCKER_GID:-999}`, então defini-lo no `.env` *chega* ao build. |
| `NEXT_PUBLIC_API_BASE` | `http://localhost:8080/api` | Endereço que as route handlers do Next.js usam para falar com a API Go. O `next build` embute valores `NEXT_PUBLIC_*` no bundle compilado, então um `ENV` de runtime é um no-op silencioso. Use `docker build --build-arg NEXT_PUBLIC_API_BASE=…` (ou `build.args` no compose). Só é necessário se API e UI rodarem em containers separados. |

### CORS na prática

A UI embutida fala com a API através das route handlers do Next.js na própria
origem, então o padrão (`CORS_ORIGIN` vazio) está correto para o deploy padrão
com docker-compose. Só configure quando um navegador em **outra** origem
precisar chamar a API diretamente:

```bash
CORS_ORIGIN=https://ark.example.com,https://admin.example.com
```

A comparação é exata — esquema, host e porta, sem barra final e sem curingas.
Uma origem que casa é devolvida junto com
`Access-Control-Allow-Credentials: true`, então mantenha a lista mínima.

## Breaking changes

Se você está atualizando um deploy existente, três mudanças recentes exigem ação:

1. **`admin_password` não é mais retornado por `GET /api/servers` nem
   `GET /api/servers/:id`.** O campo foi removido do payload de resposta para
   não entregar credenciais RCON em toda listagem. Use o endpoint dedicado
   `GET /api/servers/:id/rcon`. Qualquer cliente que lia `admin_password` da
   listagem precisa ser atualizado.

2. **`CORS_ORIGIN` não tem mais `*` como padrão.** O padrão agora é vazio —
   apenas same-origin, sem emitir nenhum header `Access-Control-Allow-Origin` —
   e o valor é uma allowlist de origens exatas separadas por vírgula, não um
   valor único nem curinga. Clientes cross-origin que funcionavam
   implicitamente passam a ser bloqueados até que a origem seja listada.

3. **`TRUSTED_PROXIES` passa a não confiar em nada por padrão.** O Gin não honra
   mais `X-Forwarded-For` de nenhuma fonte, então IPs de auditoria e rate limit
   usam o endereço real do socket. Se houver um proxy reverso na frente,
   defina `TRUSTED_PROXIES` com o IP/CIDR dele — caso contrário todas as
   requisições parecerão vir do proxy.

Endurecimento relacionado na mesma série: o container roda como usuário
não-root `arkcommander` (daí o `DOCKER_GID`) e `privileged: true` foi removido
do `docker-compose.yml`.

## Desenvolvimento

```bash
# Backend
cd server
export JWT_SECRET="$(openssl rand -base64 48)"
export LOG_FORMAT=console     # logs legíveis durante o desenvolvimento
go run main.go

# Frontend
cd ui
npm install
npm run dev
```

Alvos úteis do `make`: `make build`, `make test`, `make lint`, `make fmt`.

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
| GET | `/servers/:id/rcon` | Info RCON (único endpoint que retorna `admin_password`) |
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

Documentação interativa: `http://localhost:8080/swagger/index.html`

## Estrutura

```
├── server/               # Backend Go
│   ├── config/           # Configuração (JWT, env)
│   ├── controllers/      # Handlers HTTP
│   ├── database/         # SQLite + migrations
│   ├── middleware/       # Auth, audit, rate limit
│   ├── models/           # GORM models
│   ├── routes/           # Registro de rotas
│   ├── service/          # Lógica de negócio
│   │   ├── docker_manager/  # Docker SDK
│   │   ├── server/          # CRUD servidores
│   │   └── update/          # Monitor de updates
│   ├── utils/            # Helpers (INI, JWT, log, erros)
│   └── websocket/        # WebSocket hub
├── ui/                   # Frontend Next.js
│   ├── messages/         # Catálogos i18n (en, zh, pt-BR)
│   └── src/
│       ├── app/          # Páginas (auth + protected)
│       └── components/   # Componentes React
├── .env.example          # Template de ambiente
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

O suporte multi-idioma (en, zh, pt-BR) **já foi entregue** — veja
[Internacionalização](#internacionalização).

## Documentação

- [CHANGELOG.md](CHANGELOG.md) — histórico de versões
- [CONTRIBUTING.md](CONTRIBUTING.md) — como contribuir
- [SECURITY.md](SECURITY.md) — reporte de vulnerabilidades e notas de hardening
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) — código de conduta

## Traduções

- [English](README.md)
- [中文](README-zh.md)

## Licença

MIT
