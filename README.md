# 🔍 Shodan@Home

A self-hosted network scanner and monitoring tool inspired by Shodan. Built with Go, PostgreSQL, and React.

![Dashboard Preview](docs/dashboard.png)

## 🎯 Features

- **Automated Network Scanning**: Schedule scans with cron expressions
- **Service Detection**: Identify open ports, services, and versions
- **OS Fingerprinting**: Detect operating systems running on devices
- **Risk Assessment**: Automatically flag dangerous open ports
- **Real-time Dashboard**: Beautiful visualizations of your network
- **Historical Data**: Track changes over time
- **REST API**: Full API access to all data
- **Docker Ready**: Complete Docker setup included

## 🏗️ Architecture

```
┌─────────────┐
│  Frontend   │  React + Tailwind + Recharts
│  (Port 3000)│
└──────┬──────┘
       │
┌──────▼──────┐
│  API Server │  Go + Fiber
│  (Port 8080)│
└──────┬──────┘
       │
┌──────▼──────┐
│ PostgreSQL  │  Database
│  (Port 5432)│
└──────┬──────┘
       │
┌──────▼──────┐
│   Scanner   │  Go + Nmap (Scheduler)
└─────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Nmap installed on host (for scanner container)
- 2GB RAM minimum
- Linux/macOS (Windows WSL2 supported)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/shodan-home.git
cd shodan-home
```

2. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your network range
```

3. **Start all services**
```bash
make up
```

4. **Access the dashboard**
- Frontend: http://localhost:3000
- API: http://localhost:8080
- Health check: http://localhost:8080/health

## 📦 Available Commands

```bash
make build          # Build all containers
make up             # Start all services
make down           # Stop all services
make logs           # View all logs
make logs-api       # View API logs
make logs-scanner   # View scanner logs
make clean          # Remove everything
make scan           # Run manual scan
make stats          # Show network stats
make db-shell       # Access PostgreSQL shell
```

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SCAN_TARGET` | `192.168.1.0/24` | Network to scan (CIDR notation) |
| `SCAN_SCHEDULE` | `0 3 * * *` | Cron expression (3 AM daily) |
| `WORKERS` | `10` | Concurrent scan workers |
| `DATABASE_URL` | `postgres://...` | PostgreSQL connection string |
| `PORT` | `8080` | API server port |

### Scan Schedule Examples

```bash
# Every day at 3 AM
SCAN_SCHEDULE="0 3 * * *"

# Every 6 hours
SCAN_SCHEDULE="0 */6 * * *"

# Every Monday at midnight
SCAN_SCHEDULE="0 0 * * 1"

# Every 30 minutes (for testing)
SCAN_SCHEDULE="*/30 * * * *"
```

## 📊 API Endpoints

### Devices
```bash
GET /api/v1/devices              # List all devices
GET /api/v1/devices/:ip          # Get device details
GET /api/v1/devices/:ip/ports    # Get device ports
GET /api/v1/devices?service=ssh  # Filter by service
GET /api/v1/devices?active=true  # Only active devices
```

### Statistics
```bash
GET /api/v1/stats                # Network statistics
GET /api/v1/search?q=192.168     # Search devices
GET /api/v1/alerts               # Get unread alerts
```

### Examples

```bash
# Get all devices
curl http://localhost:8080/api/v1/devices | jq '.'

# Find SSH servers
curl http://localhost:8080/api/v1/devices?service=ssh | jq '.'

# Get network stats
curl http://localhost:8080/api/v1/stats | jq '.'
```

## 🛡️ Security Features

### Risk Detection

Automatically detects and flags:
- **Critical**: Telnet (23), FTP (21)
- **High**: SMB (445), RDP (3389), MySQL (3306), MSSQL (1433)
- **Medium**: PostgreSQL (5432), Redis (6379)

### Alerts

The system creates alerts for:
- New devices detected
- Risky ports opened
- Service version changes
- Devices going offline

## 🔨 Development

### Local Development (without Docker)

1. **Start PostgreSQL**
```bash
docker-compose up -d postgres
```

2. **Run API**
```bash
make dev-api
```

3. **Run Scanner**
```bash
make dev-scanner
```

4. **Run Frontend**
```bash
cd web && npm install && npm run dev
```

### Build Binaries

```bash
make build-all        # Build all binaries
./bin/api             # Run API
./bin/scanner         # Run scanner
./bin/scheduler       # Run scheduler
```

### Testing

```bash
make test             # Run tests
make test-coverage    # Tests with coverage
```

## 📁 Project Structure

```
shodan-home/
├── cmd/
│   ├── api/          # API server
│   ├── scanner/      # Scanner CLI
│   └── scheduler/    # Cron scheduler
├── internal/
│   ├── api/          # API handlers
│   ├── database/     # Database layer
│   └── scanner/      # Scan logic
├── web/              # React frontend
├── migrations/       # SQL migrations
├── docker-compose.yml
├── Makefile
└── README.md
```

## 🐛 Troubleshooting

### Scanner can't access network
```bash
# Ensure network_mode: host in docker-compose.yml
# Or run scanner directly on host:
make dev-scanner
```

### Permission denied for Nmap
```bash
# Scanner needs NET_ADMIN and NET_RAW capabilities
# Already configured in docker-compose.yml
```

### Database connection failed
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# View logs
docker-compose logs postgres
```

## 🚦 Roadmap

- [ ] CVE Detection (match versions with vulnerability databases)
- [ ] Email/Slack notifications
- [ ] Network topology visualization
- [ ] Custom scan profiles
- [ ] Export reports (PDF/CSV)
- [ ] Integration with Censys/VirusTotal APIs



**⭐ Star this repo if you find it useful!**