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
- CI/CD com GitHub Actions (build + test backend e frontend)
- Detecção de conflito de portas entre servidores
- Função unificada para nomeação de volumes Docker
- Internacionalização (zh-CN + en)
- Documentação Swagger/OpenAPI
- Configuração Docker multi-estágio
- Suporte a rollback em operações Docker

### Corrigido
- Nomeação de volumes de plugins agora usa função unificada `GetServerPluginsVolumeName`
- Portas duplicadas entre servidores são detectadas antes da criação
- INI content validation aplicado tanto em create quanto update

### Segurança
- Senhas armazenadas com bcrypt
- JWT com expiração e blacklist
- Refresh tokens com 30 dias de validade
- Middleware de autenticação em todas as rotas protegidas
- Auditoria de operações (create, delete, start, stop, etc.)

### Infraestrutura
- Docker multi-estágio (frontend Node 24 → backend Go 1.24 → runtime Alpine)
- Docker Compose com volume persistente e socket Docker
- GitHub Actions CI (push e PR para main)

## [0.1.0] - 2026-07-21

### Adicionado
- Projeto inicial ARK Server Commander
- Funcionalidades base de gerenciamento de servidores ARK
- Integração Docker para containers de servidor
- Interface web Next.js
- Sistema de autenticação
