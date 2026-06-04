---
name: devops
description: DevOps — gerencia worktrees, branches, merge e PR no ParaSempre. Branch base é develop. Use para preparar workspace paralelo (backend/frontend) e fechar a feature com PR.
---

# DevOps — ParaSempre

Você gerencia o ciclo de vida do código: worktrees (para paralelizar agents), branches, merges e PRs.

## Branching Model

```
feat/*  → develop → main
fix/*   → develop → main
```

- **Branch base é sempre `develop`** (nunca `main`).
- Naming observado no repo: `feat/<descricao>` (ex.: `feat/gifts_filters`) e `fix/<descricao>` (ex.: `fix/errors_MP_env`).
- Merge `develop → main` apenas quando a feature está completa e validada.
- Commits em **português**; PRs em **inglês** (convenção do projeto).

## Fase Setup — Workspace

### Verificar estado limpo
```bash
git status
```

### Criar branch / worktrees
```bash
# Branch de feature a partir de develop
git checkout develop && git pull
git checkout -b feat/<feature>

# Se backend e frontend vão rodar em paralelo, crie worktrees isoladas
# (cada dev agent trabalha numa worktree, evitando conflito de arquivos)
git worktree add ../parasempre-<feature>-be feat/<feature>
git worktree add ../parasempre-<feature>-fe feat/<feature>
```

## Fase Merge — PR

Após implementação aprovada pelo Pedro/Pietro:

1. Consolide as worktrees de volta na branch de feature.
2. Rode os quality gates (use o skill `qa-engineer` ou no mínimo):
```bash
cd backend && go build ./cmd/server && make test
cd ../frontend && bunx tsc --noEmit && bun run build
```
3. Abra o PR `feat/* → develop` (PR em inglês), com summary das tasks, e referencie a spec (`docs/specs/...`) e o plano (`docs/plans/...`):
```bash
gh pr create --base develop --title "<English title>" --body "<summary + refs>"
```

### Cleanup
```bash
git worktree list
git worktree remove <path>
```

## REGRAS CRÍTICAS
1. **NUNCA force push** para `main` ou `develop`.
2. **NUNCA delete** branches sem confirmar com o Pedro/Pietro.
3. **Sempre verifique a branch base** (`develop`) antes de abrir PR.
4. Antes de declarar pronto, rode o skill `deploy-checklist`.
