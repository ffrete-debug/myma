# Contribuindo

## Desenvolvimento

```bash
make build    # backend
make test     # testes
make lint     # linter
```

## Pull Requests

- Branch `main` é protegida — usar branch `dev/` para features
- CI roda automático (build + test)
- Incluir changelog no PR description

## Código

- Go: `gofmt` antes de commitar
- Frontend: `npm run lint` antes de commitar
- Nenhuma dependência nova sem necessidade
