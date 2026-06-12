---
description: Fluxo enxuto do ParaSempre para mudanças pequenas — vai direto ao plano e implementação, com 1 gate.
argument-hint: <descrição da mudança pequena, ex.: "adicionar filtro por preço na lista de presentes">
---

Mudança pequena/bem definida — fluxo enxuto, **sem a etapa de PO**:

**Mudança:** $ARGUMENTS

## Contexto
Consulte os skills `backend`/`frontend` para as convenções. Leia `ESCOPO-PARASEMPRE.md` apenas se precisar de contexto de produto.

## Regras
- Os agents `backend-dev`/`frontend-dev` **não commitam** — eu valido e commito.
- Tabela nova = migration com RLS (skill `migration-helper`).

## Fluxo

### 1. Plano curto  🔵 GATE
Use o skill `architect` para um plano enxuto (Chunks → tasks `[B]`/`[F]`). Para mudança trivial, pode ser uma lista curta de passos.
→ **PARE.** Apresente o plano e peça aprovação.

### 2. Implementação
Despache `backend-dev` e/ou `frontend-dev` conforme o plano, em TDD.

### 3. QA
Use o skill `qa-engineer` para rodar build/typecheck/testes (e mobile 375px se mexeu em UI).

### 4. Review
Rode `/code-review` (skill nativo) no diff. Acione `security-reviewer`/`ui-reviewer` só se realmente fizer sentido.

Ao final, peça que eu faça o commit. Comece pelo passo 1.
