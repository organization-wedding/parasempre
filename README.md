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

## Estrutura de Agentes de IA

Este projeto usa `.agents/` como diretorio padrao para configuracoes de IA:

```
.agents/
  AGENTS.md              # Instrucoes gerais do monorepo
  settings.local.json    # Permissoes locais (gitignored)
  skills/
    backend/SKILL.md     # Instrucoes especificas do backend Go
    frontend/SKILL.md    # Instrucoes especificas do frontend React/Bun

.claude -> .agents/      # Symlink para compatibilidade com Claude
```

Compativel com: **opencode**, **codex**, **claude** (via symlink), e outros agentes.

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
