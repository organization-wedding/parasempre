---
name: deploy-checklist
description: Checklist final do ParaSempre antes de declarar uma branch/PR pronta — migrations, builds, env vars, CI e testes. Use antes de abrir/mergear PR.
---

# Deploy Checklist — ParaSempre

Execute ANTES de declarar qualquer branch/PR como "pronta para merge". Se um item falhar, corrija antes de prosseguir.

### 1. Migrations
```bash
cd backend && make migrate   # contra o DB alvo (teste)
```
- Toda migration com `CREATE TABLE` tem `ENABLE ROW LEVEL SECURITY` (o job `migration-rls-check` do CI valida).

### 2. Backend — Build + Testes
```bash
cd backend && go build ./cmd/server && make test
# se mexeu em DB/migrations:
make test-integration
```

### 3. Frontend — Typecheck + Build
```bash
cd frontend && bunx tsc --noEmit && bun run build
```

### 4. Env Vars
- `backend/.env.example` e `deploy/env.example.prod` / `deploy/env.example.teste` contêm TODAS as vars novas.
- Conferir contra o `.env` local:
```bash
diff <(grep -oE '^[A-Z_]+' backend/.env.example | sort) <(grep -oE '^[A-Z_]+' .env | sort)
```

### 5. CI
```bash
gh run list --limit 5
```
- Último run do PR deve estar verde (jobs: Backend, Backend Integration, Migration RLS Check, Frontend).

### 6. Frontend — Mobile (se aplicável)
- Viewport 375px no DevTools: sem scroll horizontal; botões/inputs acessíveis.

### 7. Limpeza
- Sem `console.log` / debug solto. Sem `t.Skip`/`.only`. Sem TODO não registrado no `BACKLOG.md`.

### 8. Deploy (referência)
- Deploy é via Docker Compose + Traefik (`deploy/`): ambientes `prod` (`/opt/parasempre/prod`) e `teste` (`/opt/parasempre/teste`).
- Conferir `deploy/docker-compose.prod.yml` / `deploy/docker-compose.teste.yml` se a feature exigir nova var/serviço.

## Output

```
## Deploy Checklist — [branch]
- [x/!] Migrations + RLS: OK / pendente
- [x/!] Backend build+test: OK / falhou
- [x/!] Integration tests: OK / N/A
- [x/!] Frontend typecheck+build: OK / falhou
- [x/!] Env vars (.env.example/deploy): OK / faltando [vars]
- [x/!] CI: verde / vermelho
- [x/!] Mobile 375px: OK / N/A
- [x/!] Limpeza: OK / pendente
```
