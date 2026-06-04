# ParaSempre — Plano de Desenvolvimento com Time de IA

> Este documento descreve **como** desenvolvemos o ParaSempre usando um time de agentes de IA, e **o que** está planejado (sprints/roadmap). A arquitetura do time está completa; a seção de sprints é um template a ser preenchido com o roadmap real (ver "O que falta preencher").
>
> Estrutura espelhada e adaptada do projeto de referência `vpreis-management` (stack Next.js/Prisma/tRPC) para a stack do ParaSempre (Go + React/Bun).

---

## 1. Time de IA — Arquitetura Híbrida

Dois tipos de "peça": **agents** (subprocessos isolados) e **skills** (workflows da sessão principal).

### Agents (`.agents/agents/`) — subprocessos paralelos, contexto isolado

| Agent | Papel |
|-------|-------|
| `backend-dev` | Implementa backend Go (handler→service→repository, pgx, migrations, table-driven tests). TDD. Nunca commita. |
| `frontend-dev` | Implementa frontend React/Bun (TanStack Router, React Query, RHF+Zod, Tailwind). Mobile-first, TDD. Nunca commita. |
| `security-reviewer` | Revisa auth/payments/PII/RLS/CORS. Acionado em mudanças sensíveis. |
| `ui-reviewer` | Revisa acessibilidade, responsividade e consistência de UI. |

### Skills (`.agents/skills/`) — workflows invocados pela sessão principal

| Skill | Papel |
|-------|-------|
| `po-analyst` | Requisitos → design spec (`docs/specs/`). |
| `architect` | Spec → plano com tasks `[B]`/`[F]` (`docs/plans/`). |
| `qa-engineer` | Quality gates (build, testes, typecheck, mobile). |
| `code-reviewer` | Review 2 estágios (compliance + qualidade). |
| `devops` | Worktrees, branches, merge, PR. |
| `deploy-checklist` | Checklist final antes do merge. |
| `migration-helper` | Migrations SQL seguras (RLS, backward-compat). |
| `backend` / `frontend` | Skills de **conhecimento** (convenções de stack). |

### Por que Agents vs Skills?
- **Agent** = trabalho pesado, isolado e paralelizável (implementar, revisar a fundo). Roda em contexto próprio, não polui a sessão principal, pode rodar em worktree.
- **Skill** = workflow que a sessão principal conduz (planejar, validar gates, operar git). Roda no contexto principal, onde as decisões e gates acontecem.

---

## 2. Stack Tecnológica

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go 1.26, `net/http` stdlib (ServeMux method routing), pgx/v5 (sem ORM) |
| Frontend | React 19, Bun, TanStack Router + Query, React Hook Form + Zod, Tailwind 4 |
| Banco | PostgreSQL (Supabase), migrations SQL versionadas, RLS obrigatório |
| Integrações | MercadoPago (pagamentos), WhatsApp/OTP (auth), Firecrawl (scraping), Supabase Storage (mídia) |
| CI/CD | GitHub Actions (`ci.yml`), Docker Compose + Traefik (`deploy/`) |

---

## 3. Fluxo por Feature

```
1. PO        → po-analyst gera spec        → docs/specs/   → [GATE: aprovação]
2. Architect → architect gera plano [B]/[F] → docs/plans/   → [GATE: aprovação]
3. DevOps    → devops cria branch/worktrees (base: develop)
4. Dev       → backend-dev ∥ frontend-dev implementam (TDD, paralelo, NUNCA commitam)
5. QA        → qa-engineer roda os quality gates
6. Review    → code-reviewer (+ security-reviewer / ui-reviewer)  → [GATE: aprovação]
7. DevOps    → devops abre PR (feat/* → develop)
8. Deploy    → deploy-checklist antes do merge
```

Os **GATES** são pontos onde o Pedro/Pietro aprova antes de avançar. Os agents de implementação **nunca commitam** — o humano valida e commita.

---

## 4. Quality Gates (resumo — detalhe no skill `qa-engineer`)

- Backend: `go build ./cmd/server`, `make test`, `make test-integration`, RLS nas migrations.
- Frontend: `bunx tsc --noEmit`, `bun test`, `bun run build`, mobile 375px.
- CI verde (`Backend`, `Backend Integration`, `Migration RLS Check`, `Frontend`).

---

## 5. Branching

```
feat/*  → develop → main
fix/*   → develop → main
```
Base sempre `develop`. Commits em português, PRs em inglês.

---

## 6. Configuração (`.agents/` + `.mcp.json`)

- `.agents/settings.json` — permissions (Go/Bun) + hook `PreToolUse` que bloqueia edição de `.env`/segredos/lockfiles.
- `.mcp.json` — MCP servers: `context7` (docs de libs) e `chrome-devtools` (verificação mobile).
- **MCP opcional (não incluído):** um servidor Supabase/Postgres MCP ajudaria inspeção de DB, mas exige credenciais em env — avaliar adicionar depois.

---

## 7. Sprints / Roadmap

Ordem de prioridade definida com o casal. (Sem datas fixas — preenchemos conforme forem agendadas.)

### Visão geral
| # | Tema | Módulo | Status |
|---|------|--------|--------|
| — | Filtros de presentes | `gift` | em andamento (`feat/gifts_filters`) |
| 1 | Recados/mensagens dos presentes | `giftmessage` | próximo |
| 2 | Informações do casamento (dress code + história do casal) | novo | planejado |
| 3 | Melhorias no RSVP | `guest` | planejado |
| 4 | Fallback de login (OTP WhatsApp falhou) | `auth` | planejado |

### Detalhe por sprint

#### Sprint 1 — Recados/mensagens dos presentes (`giftmessage`)
- **Objetivo:** evoluir o recado que o convidado deixa após comprar (texto + mídia: imagem/áudio/vídeo).
- **Módulos:** `giftmessage` (backend + Supabase Storage), frontend (`GiftMessageForm`, `GiftMessageView`, `AdminGiftMessagesPage`).
- **Critérios de pronto:** a definir na spec (`po-analyst`).

#### Sprint 2 — Informações do casamento
- **Objetivo:** página(s) com **dress code** e **história do casal** (+ infos como local, horário, cronograma, mapa).
- **Módulos:** novo (provavelmente conteúdo no frontend; avaliar se precisa de backend/CMS leve).
- **Decisão em aberto:** conteúdo estático no front vs. editável pelo casal via admin. → registrar no `BACKLOG.md`.

#### Sprint 3 — Melhorias no RSVP
- **Objetivo:** aprimorar o fluxo de confirmação (livro de assinaturas, confirmação de família).
- **Módulos:** `guest` (backend), `components/rsvp/` (frontend).

#### Sprint 4 — Fallback de login
- **Objetivo:** caminho alternativo de autenticação quando o OTP por WhatsApp falha (ex.: reenviar, canal alternativo, código por outro meio).
- **Módulos:** `auth`.
- **Decisão em aberto:** qual o canal/mecanismo de fallback. → registrar no `BACKLOG.md`.

---

## O que falta preencher
- [ ] Critérios de pronto de cada sprint — definir na spec (`po-analyst`) ao iniciar cada uma.
- [ ] Sprint 2: conteúdo estático vs. editável pelo casal.
- [ ] Sprint 4: canal/mecanismo do fallback de login.
- [ ] Decisão sobre incluir MCP Supabase/Postgres.
