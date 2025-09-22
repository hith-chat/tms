# Hith Project Makefile
# This Makefile provides convenient commands for development, testing, and deployment

# Configuration
PROJECT_NAME := tms
BACKEND_DIR := ./app/backend
FRONTEND_DIR := ./app/frontend
KB_AI_SERVICE_DIR := ./app/kb-ai-service
DEPLOY_DIR := ./deploy

# Database configuration (from config.yaml)
DB_HOST := localhost
DB_PORT := 5432
DB_USER := tms
DB_PASSWORD := tms123
DB_NAME := tms

# Docker configuration
DOCKER_COMPOSE_FILE := $(DEPLOY_DIR)/docker-compose.yml

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
MAGENTA := \033[0;35m
CYAN := \033[0;36m
WHITE := \033[0;37m
NC := \033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(CYAN)Hith Project Makefile$(NC)"
	@echo "$(YELLOW)Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Development Commands
.PHONY: dev
dev: ## Start development environment
	@echo "$(BLUE)Starting development environment...$(NC)"
	@$(MAKE) docker-up
	@echo "$(GREEN)Development environment started!$(NC)"

.PHONY: dev-backend
dev-backend: ## Start backend development server
	@echo "$(BLUE)Starting backend development server...$(NC)"
	cd $(BACKEND_DIR) && go run ./cmd/api/main.go

.PHONY: dev-frontend
dev-frontend: ## Start frontend development server
	@echo "$(BLUE)Starting frontend development server...$(NC)"
	cd $(FRONTEND_DIR) && pnpm dev

.PHONY: install
install: ## Install all dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	@$(MAKE) install-backend
	@$(MAKE) install-frontend
	@echo "$(GREEN)All dependencies installed!$(NC)"

.PHONY: install-backend
install-backend: ## Install Go dependencies
	@echo "$(BLUE)Installing Go dependencies...$(NC)"
	cd $(BACKEND_DIR) && go mod download && go mod tidy

.PHONY: install-frontend
install-frontend: ## Install Node.js dependencies
	@echo "$(BLUE)Installing Node.js dependencies...$(NC)"
	cd $(FRONTEND_DIR) && pnpm install

## Build Commands
.PHONY: build
build: ## Build all components
	@echo "$(BLUE)Building all components...$(NC)"
	@$(MAKE) build-backend
	@$(MAKE) build-frontend
	@echo "$(GREEN)All components built successfully!$(NC)"

.PHONY: build-backend
build-backend: ## Build backend binary
	@echo "$(BLUE)Building backend...$(NC)"
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api/main.go

.PHONY: build-frontend
build-frontend: ## Build frontend applications
	@echo "$(BLUE)Building frontend...$(NC)"
	cd $(FRONTEND_DIR) && pnpm build

.PHONY: build-docker
build-docker: ## Build Docker images
	@echo "$(BLUE)Building Docker images...$(NC)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) build

## Test Commands
.PHONY: test
test: ## Run all tests
	@echo "$(BLUE)Running all tests...$(NC)"
	@$(MAKE) test-backend
	@$(MAKE) test-frontend
	@echo "$(GREEN)All tests completed!$(NC)"

.PHONY: test-backend
test-backend: ## Run backend tests
	@echo "$(BLUE)Running backend tests...$(NC)"
	cd $(BACKEND_DIR) && go test -v ./...

.PHONY: test-frontend
test-frontend: ## Run frontend tests
	@echo "$(BLUE)Running frontend tests...$(NC)"
	cd $(FRONTEND_DIR) && pnpm test

## Database Commands
.PHONY: db-dump
db-dump: ## Create database dump with timestamp
	@echo "$(BLUE)Creating database dump...$(NC)"
	@mkdir -p backups
	@docker run --rm --network host -e PGPASSWORD=$(DB_PASSWORD) postgres:15 pg_dump -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) > backups/$(DB_NAME)_dump_$$(date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Database dump created in backups/ directory$(NC)"

.PHONY: db-dump-schema
db-dump-schema: ## Create schema-only database dump
	@echo "$(BLUE)Creating schema-only database dump...$(NC)"
	@mkdir -p backups
	@docker run --rm --network host -e PGPASSWORD=$(DB_PASSWORD) postgres:15 pg_dump -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) --schema-only > backups/$(DB_NAME)_schema_$$(date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Schema dump created in backups/ directory$(NC)"

.PHONY: db-dump-data
db-dump-data: ## Create data-only database dump
	@echo "$(BLUE)Creating data-only database dump...$(NC)"
	@mkdir -p backups
	@docker run --rm --network host -e PGPASSWORD=$(DB_PASSWORD) postgres:15 pg_dump -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) --data-only > backups/$(DB_NAME)_data_$$(date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)Data dump created in backups/ directory$(NC)"

.PHONY: db-restore
db-restore: ## Restore database from dump file (usage: make db-restore FILE=path/to/dump.sql)
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)Error: Please specify a dump file. Usage: make db-restore FILE=path/to/dump.sql$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Restoring database from $(FILE)...$(NC)"
	@docker run --rm --network host -v $(PWD):/workspace -e PGPASSWORD=$(DB_PASSWORD) postgres:15 psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f /workspace/$(FILE)
	@echo "$(GREEN)Database restored successfully!$(NC)"

.PHONY: db-restore-direct
db-restore-direct: ## Restore database from dump file without Docker (usage: make db-restore-direct FILE=path/to/dump.sql)
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)Error: Please specify a dump file. Usage: make db-restore-direct FILE=path/to/dump.sql$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Restoring database directly from $(FILE)...$(NC)"
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $(FILE)
	@echo "$(GREEN)Database restored successfully!$(NC)"

.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "$(BLUE)Running database migrations...$(NC)"
	cd $(BACKEND_DIR) && go run ./cmd/migrate/main.go

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will drop all data)
	@echo "$(RED)WARNING: This will drop all data in the database!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "$(BLUE)Resetting database...$(NC)"; \
		docker run --rm --network host -e PGPASSWORD=$(DB_PASSWORD) postgres:15 psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"; \
		docker run --rm --network host -e PGPASSWORD=$(DB_PASSWORD) postgres:15 psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_NAME);"; \
		$(MAKE) db-migrate; \
		echo "$(GREEN)Database reset completed!$(NC)"; \
	else \
		echo "$(YELLOW)Database reset cancelled.$(NC)"; \
	fi

## Docker Commands
.PHONY: docker-up
docker-up: ## Start Docker services
	@echo "$(BLUE)Starting Docker services...$(NC)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "$(GREEN)Docker services started!$(NC)"

.PHONY: docker-down
docker-down: ## Stop Docker services
	@echo "$(BLUE)Stopping Docker services...$(NC)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "$(GREEN)Docker services stopped!$(NC)"

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

.PHONY: docker-ps
docker-ps: ## Show running Docker containers
	docker-compose -f $(DOCKER_COMPOSE_FILE) ps

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	@echo "$(BLUE)Cleaning Docker resources...$(NC)"
	docker-compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans
	docker system prune -f
	@echo "$(GREEN)Docker resources cleaned!$(NC)"

## Lint and Format Commands
.PHONY: lint
lint: ## Run linters for all components
	@echo "$(BLUE)Running linters...$(NC)"
	@$(MAKE) lint-backend
	@$(MAKE) lint-frontend
	@echo "$(GREEN)Linting completed!$(NC)"

.PHONY: lint-backend
lint-backend: ## Run Go linter
	@echo "$(BLUE)Running Go linter...$(NC)"
	cd $(BACKEND_DIR) && golangci-lint run

.PHONY: lint-frontend
lint-frontend: ## Run frontend linters
	@echo "$(BLUE)Running frontend linters...$(NC)"
	cd $(FRONTEND_DIR) && pnpm lint

.PHONY: format
format: ## Format code for all components
	@echo "$(BLUE)Formatting code...$(NC)"
	@$(MAKE) format-backend
	@$(MAKE) format-frontend
	@echo "$(GREEN)Code formatting completed!$(NC)"

.PHONY: format-backend
format-backend: ## Format Go code
	@echo "$(BLUE)Formatting Go code...$(NC)"
	cd $(BACKEND_DIR) && go fmt ./...

.PHONY: format-frontend
format-frontend: ## Format frontend code
	@echo "$(BLUE)Formatting frontend code...$(NC)"
	cd $(FRONTEND_DIR) && pnpm format

## Deployment Commands
.PHONY: deploy
deploy: ## Deploy to production
	@echo "$(BLUE)Deploying to production...$(NC)"
	@$(MAKE) build
	cd $(DEPLOY_DIR) && ./deploy.sh
	@echo "$(GREEN)Deployment completed!$(NC)"

.PHONY: deploy-staging
deploy-staging: ## Deploy to staging
	@echo "$(BLUE)Deploying to staging...$(NC)"
	@$(MAKE) build
	@echo "$(YELLOW)Staging deployment not configured yet$(NC)"

## Utility Commands
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	cd $(BACKEND_DIR) && rm -rf bin/
	cd $(FRONTEND_DIR) && rm -rf */dist/ */build/ node_modules/
	@echo "$(GREEN)Build artifacts cleaned!$(NC)"

.PHONY: logs
logs: ## Show application logs
	@echo "$(BLUE)Showing application logs...$(NC)"
	@$(MAKE) docker-logs

.PHONY: status
status: ## Show service status
	@echo "$(CYAN)Hith Service Status:$(NC)"
	@$(MAKE) docker-ps

.PHONY: backup
backup: ## Create full backup (database + files)
	@echo "$(BLUE)Creating full backup...$(NC)"
	@$(MAKE) db-dump
	@echo "$(GREEN)Full backup completed!$(NC)"

## Quick Start Commands
.PHONY: setup
setup: ## Initial project setup
	@echo "$(CYAN)Setting up Hith project...$(NC)"
	@$(MAKE) install
	@$(MAKE) docker-up
	@sleep 5
	@$(MAKE) db-migrate
	@echo "$(GREEN)Setup completed! Run 'make dev-backend' and 'make dev-frontend' to start development.$(NC)"

.PHONY: start
start: ## Start all services
	@echo "$(BLUE)Starting all Hith services...$(NC)"
	@$(MAKE) docker-up
	@echo "$(GREEN)All services started!$(NC)"

.PHONY: stop
stop: ## Stop all services
	@echo "$(BLUE)Stopping all Hith services...$(NC)"
	@$(MAKE) docker-down
	@echo "$(GREEN)All services stopped!$(NC)"

.PHONY: restart
restart: ## Restart all services
	@echo "$(BLUE)Restarting all Hith services...$(NC)"
	@$(MAKE) stop
	@$(MAKE) start
	@echo "$(GREEN)All services restarted!$(NC)"
