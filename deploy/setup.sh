#!/usr/bin/env bash
set -euo pipefail

# ParaSempre - Setup inicial da VPS
# Execute uma vez para preparar a estrutura de diretorios

DEPLOY_ROOT="/opt/parasempre"

echo "=== ParaSempre - Setup VPS ==="
echo ""

# Verificar Traefik
if ! docker ps --format '{{.Names}}' | grep -q traefik; then
  echo "AVISO: Traefik nao parece estar rodando."
  echo "Certifique-se de que o Traefik esta configurado antes de continuar."
  echo ""
fi

# Verificar rede traefik-web
if ! docker network ls --format '{{.Name}}' | grep -q traefik-web; then
  echo "ERRO: Rede 'traefik-web' nao encontrada."
  echo "Crie com: docker network create traefik-web"
  exit 1
fi

# Criar diretorios
echo "Criando diretorios..."
mkdir -p "$DEPLOY_ROOT/prod"
mkdir -p "$DEPLOY_ROOT/teste"

# Copiar templates de env
echo "Copiando templates de .env..."

if [ ! -f "$DEPLOY_ROOT/prod/.env" ]; then
  cp "$(dirname "$0")/env.example.prod" "$DEPLOY_ROOT/prod/.env"
  echo "  -> $DEPLOY_ROOT/prod/.env criado (PREENCHA OS VALORES!)"
else
  echo "  -> $DEPLOY_ROOT/prod/.env ja existe, pulando."
fi

if [ ! -f "$DEPLOY_ROOT/teste/.env" ]; then
  cp "$(dirname "$0")/env.example.teste" "$DEPLOY_ROOT/teste/.env"
  echo "  -> $DEPLOY_ROOT/teste/.env criado (PREENCHA OS VALORES!)"
else
  echo "  -> $DEPLOY_ROOT/teste/.env ja existe, pulando."
fi

echo ""
echo "=== Setup concluido ==="
echo ""
echo "Proximos passos:"
echo "  1. Preencha os .env em $DEPLOY_ROOT/prod/ e $DEPLOY_ROOT/teste/"
echo "  2. Configure os registros DNS:"
echo "     - nosparasempre.com.br       -> A -> <VPS_IP>"
echo "     - www.nosparasempre.com.br   -> A -> <VPS_IP>"
echo "     - api.nosparasempre.com.br   -> A -> <VPS_IP>"
echo "     - teste.nosparasempre.com.br -> A -> <VPS_IP>"
echo "     - api.teste.nosparasempre.com.br -> A -> <VPS_IP>"
echo "  3. Configure GitHub Environments (prod e teste) no repositorio"
echo "  4. Configure secrets SSH no GitHub (VPS_HOST, VPS_USER, VPS_SSH_KEY, VPS_SSH_PORT)"
