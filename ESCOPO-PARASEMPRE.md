# ParaSempre — Escopo do Sistema

> Escopo validado com o casal/dono do produto. Domínios no código: `guest`, `user`, `gift`, `giftmessage`, `payment`, `auth`.

## 1. Visão Geral

O ParaSempre nasceu como o **site do casamento para os convidados**: um lugar tecnológico onde eles obtêm **informações sobre o casamento**, **confirmam presença (RSVP)** e veem a **lista de presentes** dos noivos, podendo presentear online. Acoplado a isso, há um **painel administrativo** para o casal gerenciar tudo.

- **Atores:**
  - **Casal** (`groom` / `bride`) — administração completa (convidados, presentes, pagamentos, recados).
  - **Convidado** (`guest`) — vê informações, confirma presença (sua e da família), vê/compra presentes e deixa recado.
- **Autenticação:** passwordless — **telefone (BR) + OTP por WhatsApp**. `dev-login` apenas em ambiente de desenvolvimento.
- **Mobile-first é requisito formal:** toda tela deve funcionar a partir de 375px desde a primeira implementação.
- **Idioma:** UI 100% em PT-BR (com acentos corretos).
- **Escala esperada:** ~150 a 400 convidados — paginação e importação em massa importam; performance extrema não é o foco.

## 2. Módulos

### M1 — Convidados (`guest`)
- CRUD de convidados — **somente o casal** gerencia a lista completa (rotas admin protegidas por `RequireRole("groom","bride")`).
- O **convidado enxerga apenas a própria família** (`GET /api/guests/my-family`), para fins de confirmação.
- Importação em massa via CSV/XLSX.
- Agrupamento familiar (`family_group`) para confirmação em lote.
- Campos: nome, telefone (único), relacionamento (`P`/`R` — lado noivo/noiva), confirmação, grupo familiar, auditoria (RACF).

### M2 — RSVP / Confirmação de Presença
- O convidado entra via **telefone + OTP por WhatsApp** e então confirma/cancela **a própria presença e/ou a da família inteira** (grupo familiar) podendo confirmar alguns e deixar outro ou em pendentes ainda ou que outros nao vão poder ir, incluindo confirmação em lote.
- Endpoints por id, por telefone e por grupo familiar.
- Páginas: `RegisterAttendancePage`; componentes em `components/rsvp/` (livro de assinaturas, formulário de família).

### M3 — Lista de Presentes (`gift`)
- CRUD de presentes (casal), com soft-delete e deduplicação por nome.
- Importação CSV/XLSX (preview + commit).
- Scraping de produto por URL (Firecrawl) para pré-preencher cadastro.
- Lista **pública** paginada + página de detalhe.
- Campos: nome, descrição, preço (centavos), imagem, URL da loja, status.

### M4 — Pagamentos (`payment`)
- Compra de presente via MercadoPago (cartão de crédito / PIX).
- Webhook do MercadoPago com idempotência.
- Convidado vê "meus presentes/compras"; casal vê todas as transações + resumo.
- Estados: `pending`/`approved`/`rejected`/`refunded`/`cancelled`.
- **Reembolso:** manual pelo casal, caso a caso, direto no painel do MercadoPago; o sistema apenas reflete o status resultante (não dispara reembolso automático).

### M5 — Recados (`giftmessage`)
- Após comprar, o convidado deixa um recado (texto até 500 chars) com mídia opcional (imagem/áudio/vídeo) no Supabase Storage.
- Um recado por transação. O casal modera (lista + delete).

### M6 — Auth & Segurança (`auth`)
- OTP via WhatsApp (`send`/`verify`), JWT, papéis (`groom`/`bride`/`guest`).
- Rate limiting em compra, webhook e mensagens.
- Audit log de ações; RLS habilitado em todas as tabelas.

## 3. Stack
Ver `AGENTS.md` e `PLANO-DESENVOLVIMENTO.md`. Backend Go + pgx; frontend React/Bun; Postgres/Supabase; deploy Docker/Traefik.

## 4. Requisitos Não-Funcionais
- **Mobile-first** (requisito formal — 375px desde o início).
- **Idioma:** PT-BR.
- **Escala:** ~150–400 convidados; usar paginação e respeitar rate limits.
- **Segurança:** passwordless, RLS em todas as tabelas, segredos só em env, CORS restrito.

## 5. Roadmap (ordem de prioridade)
1. **Recados/mensagens dos presentes** — evoluir o módulo `giftmessage` (texto + mídia).
2. **Informações do casamento** — página(s) com **dress code** e **história do casal** (e demais infos: local, horário, cronograma, mapa).
3. **Melhorias no RSVP** — aprimorar o fluxo de confirmação (livro de assinaturas, família).
4. **Fallback de login** — alternativa de autenticação caso o OTP por WhatsApp falhe.

Detalhamento por sprint: `PLANO-DESENVOLVIMENTO.md` (Seção 7).
