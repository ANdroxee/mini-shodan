#!/usr/bin/env bash
# Installe et lance Shodan@Home sur une machine Debian 12 / Ubuntu (par exemple une VM Proxmox).
# Usage : sudo ./scripts/install.sh
set -euo pipefail

if [ "$(id -u)" -ne 0 ]; then
  echo "Lance ce script avec sudo : sudo ./scripts/install.sh" >&2
  exit 1
fi
cd "$(dirname "$0")/.."

# 1. Docker (dépôt officiel) + make
if ! command -v docker >/dev/null 2>&1; then
  echo "📦 Installation de Docker…"
  apt-get update -q
  apt-get install -y -q ca-certificates curl
  curl -fsSL https://get.docker.com | sh
fi
command -v make >/dev/null 2>&1 || apt-get install -y -q make
systemctl enable --now docker >/dev/null

# 2. Réglages : réseau détecté automatiquement, mot de passe aléatoire
if [ ! -f .env ]; then
  NET=$(ip -o -f inet route show proto kernel scope link | awk '$3 !~ /^(docker|br-|veth)/ {print $1; exit}')
  PASS=$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')
  cp .env.example .env
  sed -i "s#^SCAN_TARGET=.*#SCAN_TARGET=${NET:-192.168.1.0/24}#" .env
  sed -i "s#^POSTGRES_PASSWORD=.*#POSTGRES_PASSWORD=${PASS}#" .env
  chmod 600 .env
  echo "📝 .env créé : réseau ${NET:-192.168.1.0/24} (modifiable dans .env)"
else
  echo "📝 .env déjà présent : je le garde"
fi

# 3. Lancement (les conteneurs redémarrent tout seuls avec la VM)
echo "🚀 Construction et démarrage (quelques minutes la première fois)…"
docker compose up -d --build

IP=$(hostname -I | awk '{print $1}')
echo
echo "✅ Shodan@Home tourne. Ouvre http://${IP}:3000 depuis un appareil de la maison."
echo "   Suivre le premier scan : docker compose logs -f scanner"
