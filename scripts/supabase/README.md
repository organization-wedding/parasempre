# Scripts Supabase

Scripts SQL rodados **manualmente no SQL Editor do projeto Supabase** — fora da cadeia de migrations do app (`backend/migrations/`), porque dependem do schema `storage`, que não existe em Postgres puro (CI/local).

## Quando rodar

Ao **criar ou recriar** um ambiente (projeto Supabase) — dev, staging ou prod.

| Arquivo | O que faz |
|---------|-----------|
| `001_storage_bucket.sql` | Cria o bucket privado `gift-messages` (recados com mídia) com `file_size_limit` de 50 MB. Idempotente. |

## Checklist pós-execução

- [ ] Conferir o **limite global de upload** do projeto (Storage → Settings) ≥ 50 MB.
- [ ] Conferir que `SUPABASE_URL` e `SUPABASE_SERVICE_ROLE_KEY` (as duas) apontam para este projeto no ambiente do backend.
- [ ] Testar um upload real de mídia no recado.
