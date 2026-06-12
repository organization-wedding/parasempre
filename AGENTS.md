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

## Time de IA

Este projeto usa um **time de agentes de IA** para desenvolver features de forma estruturada. Estrutura espelhada/adaptada do projeto `vpreis-management`. Visão completa do fluxo: `PLANO-DESENVOLVIMENTO.md`.

### Agents (`.agents/agents/`) — subprocessos isolados, paralelizáveis, **nunca commitam**

| Agent | Papel |
|-------|-------|
| `backend-dev` | Implementa backend Go (handler→service→repository, pgx, migrations, table-driven tests) |
| `frontend-dev` | Implementa frontend React/Bun (TanStack Router, React Query, RHF+Zod, Tailwind) |
| `security-reviewer` | Revisa auth/payments/PII/RLS/CORS |
| `ui-reviewer` | Revisa acessibilidade, responsividade e consistência de UI |

### Skills (`.agents/skills/`) — workflows da sessão principal

| Skill | Papel |
|-------|-------|
| `po-analyst` | Requisitos → design spec (`docs/specs/`) |
| `architect` | Spec → plano com tasks `[B]`/`[F]` (`docs/plans/`) |
| `qa-engineer` | Quality gates (build, testes, typecheck, mobile) |
| `code-reviewer` | Review 2 estágios (compliance + qualidade) |
| `devops` | Worktrees, branches, merge, PR (base: `develop`) |
| `deploy-checklist` | Checklist final antes do merge |
| `migration-helper` | Migrations SQL seguras (RLS, backward-compat) |
| `backend` / `frontend` | Convenções de stack (conhecimento de domínio) |

### Fluxo por feature
`po-analyst` (spec) → `architect` (plano) → `devops` (branch/worktrees) → `backend-dev` ∥ `frontend-dev` (TDD) → `qa-engineer` (gates) → `code-reviewer` (+ security/ui) → `devops` (PR). Gates de aprovação humana entre as fases; agents de implementação não commitam.

### Documentos de referência
- `ESCOPO-PARASEMPRE.md` — escopo do produto (rascunho a validar)
- `PLANO-DESENVOLVIMENTO.md` — arquitetura do time de IA + roadmap
- `BACKLOG.md` — decisões adiadas
- `docs/specs/` e `docs/plans/` — saídas do PO e do Architect

## Convencoes Gerais

- Commits em portugues
- PRs em ingles
- Sempre rodar testes antes de commitar
- Arquivos de ambiente: `.env` (backend), variaveis no bun (frontend)

## Estrutura de Agentes

Este projeto usa `.agents/` como diretorio padrao para configuracoes de IA:

- `AGENTS.md` - Este arquivo (instrucoes gerais)
- `agents/` - Subagents (backend-dev, frontend-dev, security-reviewer, ui-reviewer)
- `skills/` - Skills de workflow e de conhecimento
- `settings.json` - Permissoes + hooks (commitado)
- `settings.local.json` - Permissoes locais (gitignored)

Config adicional na raiz: `.mcp.json` (MCP servers: context7, chrome-devtools).

Para compatibilidade com Claude, existe um symlink `.claude -> .agents`.
