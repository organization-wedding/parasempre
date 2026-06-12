---
name: migration-helper
description: Guia para criar migrations SQL seguras no ParaSempre — convenção NNN_nome.sql, BIGINT IDENTITY, RLS obrigatório, e checklist de backward-compatibility. Use ao alterar o schema do banco.
---

# Migration Helper — ParaSempre

Guia para criar e aplicar migrations **SQL puras** (sem ORM) de forma segura no Postgres/Supabase.

## Regras Fundamentais

1. Migrations são arquivos SQL versionados em `backend/migrations/`, nomeados `NNN_descricao.sql` (sequencial: a próxima após `008_create_gift_messages.sql` é `009_...`).
2. Aplicadas via `make migrate` (roda todos os arquivos na ordem). `make nuke` dropa tudo e remigra — **apenas em ambiente de teste**.
3. **NUNCA** edite uma migration já aplicada em produção — crie uma nova.
4. **RLS é obrigatório:** toda migration com `CREATE TABLE` DEVE terminar com `ALTER TABLE <name> ENABLE ROW LEVEL SECURITY;`. O job `migration-rls-check` do CI falha sem isso.

## Convenções de Schema (do código real)

```sql
CREATE TABLE IF NOT EXISTS exemplo (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nome TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    deleted_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT exemplo_nome_not_empty CHECK (length(trim(nome)) > 0),
    CONSTRAINT exemplo_status_check CHECK (status IN ('active', 'inactive')),
    CONSTRAINT exemplo_created_by_racf CHECK (created_by ~ '^[A-Z0-9]{5}$')
);

-- Unicidade considerando soft-delete: índice parcial
CREATE UNIQUE INDEX IF NOT EXISTS exemplo_nome_active_unique
    ON exemplo (nome) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS exemplo_status_idx ON exemplo (status);

ALTER TABLE exemplo ENABLE ROW LEVEL SECURITY;
```

Padrões observados no projeto:
- **PK:** sempre `BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`.
- **Timestamps:** `TIMESTAMPTZ NOT NULL DEFAULT now()`.
- **Soft delete:** `deleted_at` + `deleted_by`; índices/unicidade com `WHERE deleted_at IS NULL`.
- **Auditoria:** `created_by`/`updated_by` (RACF 5 chars, CHECK `~ '^[A-Z0-9]{5}$'`).
- **Regras de domínio no banco** via `CONSTRAINT ... CHECK` (status enum, tamanho, URL `~* '^https://'`, etc.).
- **FKs** com `REFERENCES tabela(id)` e `ON DELETE RESTRICT` quando apagar quebraria histórico (ex.: `gift_messages`).

## Workflow: Nova Migration

1. Criar `backend/migrations/NNN_descricao.sql` com o próximo número.
2. Escrever o SQL seguindo as convenções acima (lembrar do RLS!).
3. Aplicar e testar localmente:
```bash
cd backend && make migrate
make test-integration
```
4. Garantir que o `repository_postgres.go` do domínio foi atualizado (queries parametrizadas).

## Checklist de Backward-Compatibility

Antes de aplicar em produção:

- [ ] **Additive only?** Adicionar tabela/coluna/índice é seguro.
- [ ] **`CREATE TABLE` tem `ENABLE ROW LEVEL SECURITY`?** (obrigatório / CI valida)
- [ ] **Coluna `NOT NULL` nova?** Precisa de `DEFAULT`, senão falha com dados existentes.
- [ ] **Rename de coluna?** NUNCA direto — criar nova, copiar dados, remover antiga (migrations separadas).
- [ ] **Drop de coluna/tabela?** Confirmar que nenhum código em produção ainda referencia.
- [ ] **Índice em tabela grande?** Avaliar lock; considerar `CREATE INDEX CONCURRENTLY` fora de transação se necessário.
- [ ] **Mudança de enum/CHECK?** Adicionar valor é seguro; remover exige checar dados existentes.

## Output

```
## Migration Report
- Arquivo: backend/migrations/NNN_descricao.sql
- Mudança: [descrição]
- RLS habilitado: SIM / N/A (sem CREATE TABLE)
- Backward compatible: SIM / NÃO (detalhes)
- Testado (make migrate + integration): SIM / NÃO
- Pronto para produção: SIM / NÃO (bloqueios)
```
