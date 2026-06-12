# ParaSempre

Wedding Guest Management - Monorepo para gerenciamento de convidados de casamento.

## Estrutura

```
parasempre/
  backend/   # Go API (pgx + Supabase)
  frontend/  # React + Tailwind + Bun
```

## Setup Rapido

### Backend
```bash
cd backend
cp .env.example .env  # preencher variaveis
make run
```

### Frontend
```bash
cd frontend
bun install
bun dev
```

## Time de IA

Este projeto usa um **time de agentes de IA** para desenvolver features de forma estruturada (PO → Arquiteto → Devs → QA → Review), com gates de aprovação humana. Config em `.agents/` (diretorio canonico), com symlink `.claude -> .agents` para o Claude Code.

```
.agents/
  AGENTS.md              # Instrucoes gerais do monorepo (= CLAUDE.md)
  settings.json          # Permissoes + hooks (bloqueia editar .env/segredos)
  settings.local.json    # Permissoes locais (gitignored)
  agents/                # Subagents (rodam isolados, NUNCA commitam)
    backend-dev.md       #   implementa backend Go
    frontend-dev.md      #   implementa frontend React/Bun
    security-reviewer.md #   revisa auth/payments/PII/RLS
    ui-reviewer.md       #   revisa acessibilidade/responsividade
  skills/                # Workflows + conhecimento de dominio
    backend/ frontend/   #   convencoes de stack
    po-analyst/ architect/ qa-engineer/ code-reviewer/
    devops/ deploy-checklist/ migration-helper/
  commands/              # Slash commands de fluxo
    feature.md fix.md quick.md

.mcp.json                # MCP servers (context7, chrome-devtools)
.claude -> .agents/      # Symlink para compatibilidade com Claude
```

### Fluxo de desenvolvimento

Tres comandos cobrem os tipos de trabalho (cada um para nos gates pra voce aprovar):

| Comando | Para que | Branch |
|---------|----------|--------|
| `/feature <desc>` | Feature nova (spec → plano → implementacao → review) | `feat/*` |
| `/fix <bug + repro>` | Corrigir bug (causa-raiz → teste de regressao → correcao) | `fix/*` |
| `/quick <desc>` | Mudanca pequena (plano curto → implementacao) | `feat/*` |

Guia completo de uso: [`docs/COMO-USAR-O-TIME-DE-IA.md`](docs/COMO-USAR-O-TIME-DE-IA.md).
Escopo, roadmap e decisoes: [`ESCOPO-PARASEMPRE.md`](ESCOPO-PARASEMPRE.md), [`PLANO-DESENVOLVIMENTO.md`](PLANO-DESENVOLVIMENTO.md), [`BACKLOG.md`](BACKLOG.md).

Compativel com: **opencode**, **codex**, **claude** (via symlink), e outros agentes — os arquivos `AGENTS.md`/`skills` sao portaveis; `agents/` e `commands/` sao recursos do Claude Code.

### Windows: Configurando Symlinks

O arquivo `.claude` eh um symlink para `.agents/`. No Windows, symlinks requerem configuracao adicional.

**Verificar se symlinks estao habilitados:**
```bash
git config core.symlinks
```

Se retornar `false` ou nada:

**Opcao 1 - Habilitar globalmente (recomendado):**
```bash
git config --global core.symlinks true
```

**Opcao 2 - Habilitar apenas neste repositorio:**
```bash
git config core.symlinks true
```

**Requisitos no Windows:**
- Git for Windows 2.19+ (versoes anteriores tem suporte limitado)
- Executar terminal como Administrator OU ter "Developer Mode" habilitado no Windows 10/11

**Se o symlink nao funcionar apos clonar:**
```bash
# Recriar o symlink manualmente (no Git Bash ou PowerShell como Admin)
cd parasempre
rm -rf .claude
cmd /c mklink /D .claude .agents
```

**Verificar se o symlink esta correto:**
```bash
ls -la .claude
# Deve mostrar: .claude -> .agents
```

## Stack

- **Backend**: Go 1.26, pgx/v5, net/http stdlib
- **Frontend**: React, Tailwind, Bun
- **Database**: PostgreSQL (Supabase)
