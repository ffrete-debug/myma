# Changelog

Todas as mudanças notáveis neste projeto serão documentadas aqui.

## [Unreleased]

### Adicionado
- Gerenciamento completo de servidores ARK (CRUD, start, stop, restart, recreate)
- Painel web Next.js 15 com React 19 e TypeScript
- Autenticação JWT com refresh token e blacklist
- Validação de formato INI (GameUserSettings.ini e Game.ini)
- Controle de concorrência por usuário com sync.Map
- Isolamento de estado de pull de imagem Docker
- Monitoramento de atualização com WebSocket
- Estrutura de erros padronizada (APIError)
- Audit logging de operações sensíveis
- CI/CD com GitHub Actions — job `backend` (build, `go vet`, `golangci-lint`,
  `go test`) e job `frontend` (`npm ci`, `npm audit`, typecheck `tsc --noEmit`,
  lint, `npm test` → `vitest run`, build)
- Detecção de conflito de portas entre servidores
- Função unificada para nomeação de volumes Docker
- Internacionalização via next-intl (en, zh, pt-BR), padrão `en`, seleção
  persistida no cookie `NEXT_LOCALE`
- Documentação Swagger/OpenAPI
- Configuração Docker multi-estágio
- Suporte a rollback em operações Docker
- `.env.example` — template de ambiente versionado, referenciado pelo passo
  `cp .env.example .env` do Quick Start (antes o arquivo não existia e o
  container subia em crash-loop por falta de `JWT_SECRET`)
- README em inglês como página inicial do repositório, com o conteúdo em
  português movido para `README-pt-BR.md` e cross-links entre
  `README.md` / `README-pt-BR.md` / `README-zh.md`
- Documentação das variáveis de ambiente `DB_PATH`, `SERVER_PORT`,
  `CORS_ORIGIN`, `TRUSTED_PROXIES`, `LOG_LEVEL`, `LOG_FORMAT`,
  `NEXT_PUBLIC_API_BASE` e `DOCKER_GID`

### Alterado — atenção, quebra compatibilidade
- **`admin_password` não é mais serializado em `GET /api/servers` nem em
  `GET /api/servers/:id`.** Use o endpoint dedicado `GET /api/servers/:id/rcon`
  para obter a credencial RCON. Clientes que liam o campo da listagem precisam
  ser atualizados.
- **`CORS_ORIGIN` não tem mais `*` como padrão.** O padrão passa a ser vazio
  (apenas same-origin, sem emitir `Access-Control-Allow-Origin`) e o valor agora
  é uma allowlist de origens exatas separadas por vírgula. Clientes cross-origin
  que funcionavam implicitamente passam a ser bloqueados até serem listados.
- **`TRUSTED_PROXIES` passa a não confiar em nenhum proxy por padrão.** O
  `X-Forwarded-For` deixa de ser honrado, então o IP de auditoria e o rate limit
  usam o endereço real do socket. Deploys atrás de proxy reverso precisam
  declarar o IP/CIDR do proxy.

### Corrigido
- Nomeação de volumes de plugins agora usa função unificada `GetServerPluginsVolumeName`
- Portas duplicadas entre servidores são detectadas antes da criação
- INI content validation aplicado tanto em create quanto update
- CHANGELOG: a ordem dos estágios do Docker estava invertida ("frontend →
  backend → runtime"; o build é backend → frontend → runtime) e a versão do Node
  estava errada ("Node 24"; o build usa Node 20) — ver seção Infraestrutura
- CI: a linha de CHANGELOG afirmava "build + test backend e frontend", mas
  nenhum workflow executava os testes do frontend. Os testes já existiam
  (`ui/src/lib/*.test.ts`, `npm test` → `vitest run`) e faltava apenas o step no
  job `frontend`. O step foi adicionado (junto com um step de typecheck), de modo
  que a CI agora executa esses testes de fato e a descrição acima reflete o que
  ela realmente roda
- README: bloco de roadmap duplicado removido, e "Multi-language support"
  deixou de ser listado como item pendente — o i18n já está entregue

### Segurança
- Senhas armazenadas com bcrypt
- JWT com expiração e blacklist
- Refresh tokens com 30 dias de validade
- Middleware de autenticação em todas as rotas protegidas
- Auditoria de operações (create, delete, start, stop, etc.)
- `JWT_SECRET` obrigatório na inicialização: mínimo de 32 caracteres e recusa de
  padrões fracos conhecidos; o processo encerra com instruções em vez de subir
  com uma chave insegura
- CORS por allowlist explícita em vez de `*`
- `SetTrustedProxies` configurável, com padrão de não confiar em nenhum proxy
- Container roda como usuário não-root `arkcommander` (build arg `DOCKER_GID`
  para acesso ao socket do Docker)
- `privileged: true` removido do `docker-compose.yml`
- Socket do Docker montado como `:ro` (defesa em profundidade)
- `opencode.json` removido do repositório

### Infraestrutura
- Docker multi-estágio, nesta ordem: backend Go (`golang:1.24.4-alpine`) →
  frontend Node (`node:20.19.0-alpine`) → runtime `alpine:3.22` com `nodejs`
  (22.x) e `tini`
- Imagens base fixadas em tag de patch exata, para que o SO de runtime — e a
  versão major do Node instalada pelo `apk` — não mude entre dois builds do
  mesmo commit
- Frontend é compilado em Node 20, servido em Node 22 e verificado em Node 22 na
  CI; backend em Go 1.24
- `tini` como PID 1 e `entrypoint.sh` supervisionando os dois processos, para
  propagar SIGTERM e derrubar o container se qualquer um dos dois morrer
- `NEXT_PUBLIC_API_BASE` passa a ser build arg do estágio de frontend: `next
  build` inlina valores `NEXT_PUBLIC_*` no bundle, então defini-lo como `ENV` de
  runtime era um no-op silencioso
- Docker Compose com volume persistente, `env_file: .env`, healthcheck em
  `/health` e socket Docker
- GitHub Actions CI (push e PR para main)

## [0.1.0] - 2026-07-21

### Adicionado
- Projeto inicial ARK Server Commander
- Funcionalidades base de gerenciamento de servidores ARK
- Integração Docker para containers de servidor
- Interface web Next.js
- Sistema de autenticação
