---
description: Fluxo de correção de bug do ParaSempre — diagnostica a causa-raiz, escreve teste de regressão, corrige e valida. Branch fix/*.
argument-hint: <descrição do bug + como reproduzir, ex.: "compra de presente fica 'pending' mesmo após webhook aprovar">
---

Você vai corrigir o seguinte bug usando o Time de IA do ParaSempre, **priorizando causa-raiz sobre sintoma** e parando nos gates:

**Bug:** $ARGUMENTS

## Contexto
Consulte os skills `backend`/`frontend` para as convenções. Leia `ESCOPO-PARASEMPRE.md` só se precisar entender o comportamento esperado.

## Regras de ouro
- **Corrija a causa-raiz, não o sintoma.** Nada de remendo que esconde o problema.
- **PARE nos 🔵 GATES** e espere minha aprovação.
- Os agents `backend-dev`/`frontend-dev` **não commitam** — eu valido e commito.
- Se a correção exigir mudança de schema, use o skill `migration-helper` (RLS!).

## Fluxo

### 1. Reproduzir + Diagnosticar (causa-raiz)  🔵 GATE
- Reproduza o bug (rode o app/teste, leia os logs, siga o caminho do código: handler → service → repository, ou page → query → api).
- Para investigação ampla, despache o agent `Explore` ou um dev agent em modo leitura.
- Identifique a **causa-raiz** exata (`arquivo:linha`) e por que acontece.
- Avalie o **raio de impacto**: o mesmo bug existe em outros lugares?
→ **PARE.** Apresente: (a) como reproduz, (b) causa-raiz, (c) correção proposta, (d) raio de impacto. Espere minha aprovação.

### 2. Teste de regressão (RED)
Escreva um teste que **reproduz o bug e falha** hoje (table-driven no backend, `bun test` no frontend). Esse teste é a prova de que entendemos o problema.

### 3. Correção (GREEN)
Despache o agent apropriado (`backend-dev` e/ou `frontend-dev`) para a correção **mínima na causa-raiz**, deixando o teste do passo 2 verde. Se o bug existia em vários lugares, corrija todos.

### 4. QA
Use o skill `qa-engineer`: rode build, testes (unit + integration se tocou DB), typecheck e o teste de regressão novo. Garanta que **nada mais quebrou**.

### 5. Review  🔵 GATE
Rode `/code-review` (skill nativo) no diff. Acione o agent `security-reviewer` se o bug tocar auth/payments/PII. 
→ **PARE.** Apresente os achados e espere eu decidir.

### 6. Fechamento
Use o skill `devops` para abrir o PR `fix/<descrição> → develop` (PR em inglês), referenciando a causa-raiz e o teste de regressão. **Não commite por mim** — peça que eu faça o commit/merge.

Comece pelo passo 1.
