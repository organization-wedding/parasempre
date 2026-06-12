# Backlog — Decisões Adiadas

Registre aqui decisões e itens adiados durante o desenvolvimento, para não se perderem. Cada item: o quê, por quê foi adiado, e (se houver) condição/data para retomar.

> Template — substitua/adicione conforme as decisões aparecerem. O skill `po-analyst` lê este arquivo ao iniciar uma feature.

## Formato

```
- [ ] <descrição do item> — adiado em <data>, motivo: <por quê>. Retomar quando: <condição>.
```

## Infra de IA / Config
- [ ] Decidir se adiciona MCP Supabase/Postgres ao `.mcp.json` (precisa de credenciais em env).
- [ ] Avaliar instalar o ecossistema de skills "superpowers" (hoje os skills são autocontidos).

## Backend
- [ ] **giftmessage `BucketExists` loga corpo do Supabase (até 512 bytes) no boot check** — adiado em 2026-06-04 (review da Sprint 1, risco baixo). Trim pra status code / sanitizar, evitando metadados internos do Supabase em logs agregados. Retomar quando: mexer em `supabase_storage.go` ou hardening de logs.
- [ ] **giftmessage `resolveMediaMIME`: label de erro pode sair vazio** — adiado em 2026-06-04 (cosmético). Quando o tipo declarado é vazio + bytes perigosos, a mensagem vira "tipo de mídia não suportado: " (sem tipo). Usar o `sniffed` como label nesse caso. Retomar quando: mexer em `media.go`.
- [ ] **guest confirm/cancel por ID: enumeração via 403 vs 404** — adiado em 2026-06-04 (review do fix do filtro de convidados, BAIXO/pré-existente). `setConfirmed` (service.go) responde 403 quando o ID alvo é de outra família e 404 quando não existe, permitindo a um convidado enumerar IDs válidos. Unificar para 404 genérico. Retomar quando: hardening de auth ou mexer em confirm/cancel.
- [ ] **`my-family` expõe `created_by`/`updated_by` (URACFs do casal) no JSON** — adiado em 2026-06-04 (BAIXO/pré-existente). A struct `Guest` exportada inclui os URACFs do noivo/noiva em endpoints guest-facing. Avaliar DTO separado ou `json:"-"` para convidados comuns. Retomar quando: hardening de PII ou refatorar respostas de guest.
- [ ] **`DevOnly` middleware usa denylist (`== "production"`) em vez de allowlist** — adiado em 2026-06-04 (BAIXO/defense-in-depth; proteção real é o nil-check em main.go). Inverter para allowlist de ambientes não-prod. Retomar quando: hardening de config/ambiente.

## Frontend
- (vazio)

## Produto / Escopo
- [ ] **Infos do casamento (Sprint 2):** conteúdo estático no frontend vs. editável pelo casal via admin (CMS leve). Decidir na spec.
- [ ] **Fallback de login (Sprint 4):** definir canal/mecanismo quando o OTP por WhatsApp falha (reenvio, canal alternativo, código por outro meio).
