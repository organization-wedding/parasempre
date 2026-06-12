---
description: Roda o fluxo completo do Time de IA do ParaSempre para uma feature, parando nos gates de aprovação.
argument-hint: <descrição da feature, ex.: "recados dos presentes (Sprint 1)">
---

Você vai conduzir o desenvolvimento da seguinte feature usando o Time de IA do ParaSempre, **parando nos gates** para aprovação humana:

**Feature:** $ARGUMENTS

## Contexto (leia antes de começar)
Leia `ESCOPO-PARASEMPRE.md`, `PLANO-DESENVOLVIMENTO.md` (Seção 7) e `BACKLOG.md` para entender produto, roadmap e decisões abertas. Consulte os skills `backend`/`frontend` para as convenções de stack.

## Regras de ouro
- **PARE em cada 🔵 GATE** e espere minha aprovação explícita antes de seguir. Não avance sozinho.
- Os agents de implementação (`backend-dev`, `frontend-dev`) **nunca commitam** — eu valido e commito.
- Toda tabela nova = migration com RLS (use o skill `migration-helper`).
- Mantenha contratos backend↔frontend em sincronia (schema Zod ↔ payload Go).

## Fluxo

### 1. PO — Design Spec  🔵 GATE
Use o skill `po-analyst`. Faça as perguntas de design **uma de cada vez**. Gere `docs/specs/AAAA-MM-DD-<modulo>-design.md`.
→ **PARE.** Apresente a spec e peça minha aprovação.

### 2. Architect — Plano  🔵 GATE
Use o skill `architect`. Decomponha a spec aprovada em `docs/plans/AAAA-MM-DD-<modulo>-plan.md` com tasks `[B]`/`[F]`/`[BF]`/`[D]`, identificando paralelismo.
→ **PARE.** Apresente o plano e peça minha aprovação.

### 3. DevOps — Branch
Use o skill `devops` para criar a branch a partir de `develop` (e worktrees se houver backend+frontend em paralelo).

### 4. Implementação
Despache o agent `backend-dev` nas tasks `[B]` e o `frontend-dev` nas `[F]`, em paralelo quando possível, em TDD. Se mexer no schema, use o skill `migration-helper` antes.

### 5. QA
Use o skill `qa-engineer` para rodar os quality gates. Se FAIL, volte ao passo 4 e corrija.

### 6. Review  🔵 GATE
Use o skill `code-reviewer` no diff. Acione o agent `security-reviewer` se a feature tocar auth/payments/PII/migrations, e o `ui-reviewer` se tocar UI.
→ **PARE.** Apresente os achados (CRITICAL/IMPORTANT/SUGGESTION) e espere eu decidir o que corrigir.

### 7. Fechamento
Use o skill `deploy-checklist`. Se passar, use o skill `devops` para abrir o PR `feat/* → develop` (PR em inglês). **Não commite por mim** — peça que eu faça o commit/merge final.

Comece pelo passo 1.
