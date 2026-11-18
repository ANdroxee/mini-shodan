# 🔍 Shodan@Home

Un petit Shodan pour la maison : il **scanne tout ton réseau en boucle**, repère ce qui est branché, les portes laissées ouvertes, et t'explique en français courant quoi faire.

Construit avec Go, PostgreSQL, nmap et React.

## Ce que ça fait

- **Scan en boucle** — selon un planning (par défaut tous les jours à 3 h) et à la demande depuis le bouton « Vérifier maintenant »
- **Score de sécurité** sur 100 et un résumé lisible : « Ton réseau va plutôt bien. 2 choses à régler. »
- **Carte de la maison** — tous les appareils autour de la box, colorés selon leur état (Urgent, À vérifier, Nouveau, RAS, Hors ligne)
- **Alertes intelligentes**, sans doublons d'un scan à l'autre :
  - un **appareil inconnu** apparaît (le tout premier scan sert de référence) ;
  - un **port risqué** s'ouvre — l'alerte se clôt toute seule quand il se ferme ;
  - le **logiciel** derrière un port change de version ;
  - un appareil **ne répond plus**.
- Pour chaque alerte : ce qui se passe, pourquoi c'est important, **comment corriger** étape par étape, et la preuve technique (sortie nmap)
- **Renommer** un appareil, dire « **je le connais** »
- Détection du système (OS), de l'adresse MAC et du fabricant

### Ports surveillés

| Gravité | Ports |
|---------|-------|
| Urgent (critique) | Telnet (23), FTP (21) |
| À vérifier (élevée) | SMB (445), Bureau à distance (3389), MySQL (3306), SQL Server (1433) |
| À vérifier (moyenne) | PostgreSQL (5432), Redis (6379) |

## Architecture

```
┌──────────────┐
│  Interface   │  React + Tailwind, servie par nginx (port 3000)
└──────┬───────┘  nginx relaie /api vers l'API
┌──────▼───────┐
│     API      │  Go + Fiber (port 8080, accessible en local seulement)
└──────┬───────┘
┌──────▼───────┐
│  PostgreSQL  │  schéma créé et mis à jour automatiquement
└──────▲───────┘
┌──────┴───────┐
│   Scanner    │  Go + nmap, tourne en boucle (planning + scans demandés)
└──────────────┘  réseau de l'hôte, pour voir les appareils du réseau local
```

## Démarrage

Prérequis : Docker et Docker Compose, sur une machine **Linux** branchée sur le réseau à surveiller (un Raspberry Pi ou un serveur de homelab fait très bien l'affaire).

```bash
git clone https://github.com/ANdroxee/mini-shodan.git
cd mini-shodan
cp .env.example .env      # règle SCAN_TARGET sur ton réseau et change POSTGRES_PASSWORD
make up
```

Puis ouvre **http://IP-de-la-machine:3000**. Le premier scan démarre tout de suite et prend quelques minutes.

> Sur macOS ou Windows, Docker tourne dans une machine virtuelle : le scanner ne voit pas directement le réseau de la maison. Lance-le plutôt sur une machine Linux.

## Installer dans une VM Proxmox

**1. Créer la VM** (dans l'interface Proxmox, « Create VM ») :

| Réglage | Valeur |
|---------|--------|
| ISO | Debian 12 (netinst) |
| CPU / RAM / disque | 2 cœurs · 2 Go · 16 Go |
| Réseau | **Bridge `vmbr0`** (le pont relié à ta box), modèle VirtIO |
| Pare-feu Proxmox sur la carte réseau | décoché (ou autoriser le port 3000) |

Le point important, c'est le **bridge** : la VM doit être sur le même réseau que tes appareils, comme un vrai ordinateur branché sur la box. Sans ça, le scanner ne voit ni les appareils ni leurs adresses MAC.

**2. Installer Debian** : options par défaut, cocher « serveur SSH », pas d'environnement graphique. Ensuite, donne-lui une **IP fixe** (bail DHCP réservé dans l'interface de ta box) pour toujours retrouver l'interface à la même adresse.

**3. Installer Shodan@Home** (en SSH sur la VM) :

```bash
sudo apt install -y git
git clone https://github.com/ANdroxee/mini-shodan.git
cd mini-shodan
sudo ./scripts/install.sh
```

Le script installe Docker, détecte ton réseau, crée le `.env` avec un mot de passe aléatoire et lance tout. Les conteneurs redémarrent seuls avec la VM. Pense à cocher « Start at boot » dans les options de la VM.

**Mettre à jour plus tard :** `git pull && sudo docker compose up -d --build`

## Réglages (`.env`)

| Variable | Défaut | Rôle |
|----------|--------|------|
| `SCAN_TARGET` | `192.168.1.0/24` | Réseau à scanner (CIDR) |
| `SCAN_SCHEDULE` | `0 3 * * *` | Fréquence (cron) : tous les jours à 3 h |
| `SCAN_TIMEZONE` | `Europe/Paris` | Fuseau horaire du planning |
| `WORKERS` | `10` | Appareils scannés en parallèle |
| `POSTGRES_PASSWORD` | `shodan123` | Mot de passe de la base — **à changer** |

Exemples de planning : `@every 30m` (toutes les 30 minutes), `0 */6 * * *` (toutes les 6 heures), `0 0 * * 1` (chaque lundi à minuit).

## API

```bash
GET   /api/v1/summary              # score, compteurs, dernier et prochain scan
GET   /api/v1/devices              # appareils, ports et statut
GET   /api/v1/devices/:ip
PATCH /api/v1/devices/:ip          # {"name": "Caméra du salon"} ou {"known": true}
GET   /api/v1/alerts               # alertes ouvertes (?all=true : aussi les résolues)
POST  /api/v1/alerts/:id/read      # {"read": true}
POST  /api/v1/alerts/read-all
POST  /api/v1/scans                # demander un scan immédiat
GET   /api/v1/scans/latest
GET   /api/v1/stats
GET   /api/v1/search?q=192.168
```

## Commandes utiles

```bash
make up / make down     # démarrer / arrêter
make logs-scanner       # suivre les scans
make scan               # scan ponctuel
make db-shell           # console PostgreSQL
```

### Développement sans Docker

```bash
docker compose up -d postgres
make dev-api                    # API sur :8080
make dev-scheduler              # boucle de scan (nmap installé sur la machine)
cd web && npm install && npm run dev   # interface sur :3000
```

## Structure

```
cmd/
  api/          serveur API
  scanner/      scan ponctuel
  scheduler/    boucle de scan (cron + scans demandés)
internal/
  api/          routes et handlers
  database/     accès PostgreSQL + schema.sql
  monitor/      un scan complet → base + alertes
  risk/         ports risqués et gravité
  scanner/      pilotage de nmap
web/            interface React (style « grand public »)
```

## Sécurité

- L'interface n'a pas de compte : ne l'expose **pas** sur Internet, garde-la sur ton réseau local.
- La base et l'API n'écoutent que sur la machine elle-même (`127.0.0.1`).
- Le scanner ne tourne pas en root : nmap reçoit seulement les capacités réseau nécessaires.
- Ne scanne que des réseaux qui t'appartiennent.

## Pistes

- [ ] Notifications sur le téléphone
- [ ] Détection de CVE à partir des versions
- [ ] Export de rapports
