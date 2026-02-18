# ParaSempre - Wedding Guest Management

Monorepo para gerenciamento de convidados de casamento.

## Estrutura

```
parasempre/
  backend/   # Go API (pgx + Supabase)
  frontend/  # React + Tailwind + Bun
```

## Stack

- **Backend**: Go 1.26, pgx/v5, net/http stdlib
- **Frontend**: React, Tailwind, Bun
- **Database**: PostgreSQL (Supabase)

## Comandos Principais

| Task | Backend | Frontend |
|------|---------|----------|
| Dev | `make run` | `bun dev` |
| Build | `make build` | `bun run build` |
| Test | `make test` | `bun test` |
| Typecheck | `go build` | `bunx tsc --noEmit` |

## Skills Disponiveis

- `backend` - Instrucoes especificas do backend Go
- `frontend` - Instrucoes especificas do frontend React/Bun

## Convencoes Gerais

- Commits em portugues
- PRs em ingles
- Sempre rodar testes antes de commitar
- Arquivos de ambiente: `.env` (backend), variaveis no bun (frontend)

## Estrutura de Agentes

Este projeto usa `.agents/` como diretorio padrao para configuracoes de IA:

- `AGENTS.md` - Este arquivo (instrucoes gerais)
- `skills/` - Instrucoes especificas por dominio
- `settings.local.json` - Permissoes locais (gitignored)

Para compatibilidade com Claude, existe um symlink `.claude -> .agents`.
