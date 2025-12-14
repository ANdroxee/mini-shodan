.PHONY: help build run stop clean test scan

# Variables
DOCKER_COMPOSE = docker-compose
GO = go
API_PORT = 8080
FRONTEND_PORT = 3000

help: ## Affiche l'aide
	@echo "Shodan@Home - Commandes disponibles:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build tous les containers
	$(DOCKER_COMPOSE) build

up: ## Démarre tous les services
	$(DOCKER_COMPOSE) up -d
	@echo "✅ Services started!"
	@echo "📡 API: http://localhost:$(API_PORT)"
	@echo "🌐 Frontend: http://localhost:$(FRONTEND_PORT)"

down: ## Arrête tous les services
	$(DOCKER_COMPOSE) down

stop: ## Stop les services sans les supprimer
	$(DOCKER_COMPOSE) stop

restart: ## Redémarre tous les services
	$(DOCKER_COMPOSE) restart

logs: ## Affiche les logs
	$(DOCKER_COMPOSE) logs -f

logs-api: ## Logs de l'API uniquement
	$(DOCKER_COMPOSE) logs -f api

logs-scanner: ## Logs du scanner uniquement
	$(DOCKER_COMPOSE) logs -f scanner

clean: ## Supprime tout (containers, volumes, images)
	$(DOCKER_COMPOSE) down -v --rmi all
	rm -rf postgres_data

rebuild: clean build up ## Clean + rebuild + start

# Commandes de développement local (sans Docker)
dev-deps: ## Install Go dependencies
	$(GO) mod download
	$(GO) mod tidy

dev-api: ## Lance l'API en local
	DATABASE_URL="postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable" \
	PORT=8080 \
	$(GO) run cmd/api/main.go

dev-scanner: ## Lance un scan en local
	DATABASE_URL="postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable" \
	SCAN_TARGET="192.168.1.0/24" \
	WORKERS=10 \
	$(GO) run cmd/scanner/main.go

dev-scheduler: ## Lance le scheduler en local
	DATABASE_URL="postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable" \
	SCAN_TARGET="192.168.1.0/24" \
	SCAN_SCHEDULE="*/5 * * * *" \
	WORKERS=10 \
	RUN_ON_START=true \
	$(GO) run cmd/scheduler/main.go

# Commandes de base de données
db-shell: ## Accède au shell PostgreSQL
	$(DOCKER_COMPOSE) exec postgres psql -U shodan -d shodan_home

db-migrate: ## Applique les migrations
	$(DOCKER_COMPOSE) exec postgres psql -U shodan -d shodan_home -f /docker-entrypoint-initdb.d/001_init.sql

db-reset: ## Reset la base de données
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d postgres
	@sleep 5
	$(MAKE) db-migrate

# Tests
test: ## Lance les tests Go
	$(GO) test -v ./...

test-coverage: ## Tests avec coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out

# Scan manuel
scan: ## Lance un scan manuel
	$(DOCKER_COMPOSE) exec scanner ./scanner

scan-quick: ## Scan rapide (sans version detection)
	SCAN_TARGET="192.168.1.0/24" $(DOCKER_COMPOSE) run --rm scanner ./scanner

# Build binaries locaux
build-api: ## Build l'API en local
	CGO_ENABLED=0 $(GO) build -o bin/api cmd/api/main.go

build-scanner: ## Build le scanner en local
	CGO_ENABLED=0 $(GO) build -o bin/scanner cmd/scanner/main.go

build-scheduler: ## Build le scheduler en local
	CGO_ENABLED=0 $(GO) build -o bin/scheduler cmd/scheduler/main.go

build-all: build-api build-scanner build-scheduler ## Build tous les binaires

# Stats
stats: ## Affiche les stats du réseau
	@curl -s http://localhost:$(API_PORT)/api/v1/stats | jq '.'

devices: ## Liste tous les devices
	@curl -s http://localhost:$(API_PORT)/api/v1/devices | jq '.'

# Frontend
frontend-install: ## Install frontend dependencies
	cd web && npm install

frontend-dev: ## Lance le frontend en dev
	cd web && npm run dev

frontend-build: ## Build le frontend
	cd web && npm run build