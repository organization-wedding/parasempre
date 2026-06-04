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
- (vazio)

## Frontend
- (vazio)

## Produto / Escopo
- [ ] **Infos do casamento (Sprint 2):** conteúdo estático no frontend vs. editável pelo casal via admin (CMS leve). Decidir na spec.
- [ ] **Fallback de login (Sprint 4):** definir canal/mecanismo quando o OTP por WhatsApp falha (reenvio, canal alternativo, código por outro meio).
