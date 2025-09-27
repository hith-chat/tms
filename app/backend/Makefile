# Backend Makefile for TMS Project

# Configuration
APP_NAME := tms-backend
MAIN_PATH := ./cmd/api/main.go
DOCS_DIR := ./docs
SWAGGER_OUTPUT := $(DOCS_DIR)/swagger.json
SWAGGER_YAML := $(DOCS_DIR)/swagger.yaml

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

.PHONY: help
help: ## Show this help message
	@echo "$(BLUE)TMS Backend Makefile$(NC)"
	@echo "$(YELLOW)Available commands:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Install/update dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	go mod tidy
	go mod download

.PHONY: install-swaggo
install-swaggo: ## Install swaggo/swag for API documentation
	@echo "$(BLUE)Installing swaggo/swag...$(NC)"
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: docs
docs: install-swaggo ## Generate OpenAPI documentation
	@echo "$(BLUE)Generating API documentation...$(NC)"
	@mkdir -p $(DOCS_DIR)
	swag init -g $(MAIN_PATH) -o $(DOCS_DIR) --parseDependency --parseInternal
	@echo "$(GREEN)API documentation generated in $(DOCS_DIR)$(NC)"
	@echo "$(YELLOW)Generated files:$(NC)"
	@echo "  - $(SWAGGER_OUTPUT)"
	@echo "  - $(SWAGGER_YAML)"
	@echo "  - $(DOCS_DIR)/docs.go"

.PHONY: build
build: ## Build the application
	@echo "$(BLUE)Building $(APP_NAME)...$(NC)"
	go build -o bin/$(APP_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build completed: bin/$(APP_NAME)$(NC)"

.PHONY: run
run: ## Run the application
	@echo "$(BLUE)Running $(APP_NAME)...$(NC)"
	go run $(MAIN_PATH)

.PHONY: dev
dev: ## Run the application in development mode with air
	@echo "$(BLUE)Starting development server...$(NC)"
	air

.PHONY: test
test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v ./...

.PHONY: clean
clean: ## Clean build artifacts and docs
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -rf $(DOCS_DIR)/
	@echo "$(GREEN)Cleanup completed$(NC)"

.PHONY: lint
lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	golangci-lint run

.PHONY: format
format: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...

.PHONY: docs-serve
docs-serve: docs ## Generate docs and serve them locally
	@echo "$(BLUE)Serving API documentation...$(NC)"
	@echo "$(YELLOW)API documentation will be available at: http://localhost:8080/swagger/index.html$(NC)"
	@echo "$(YELLOW)Make sure to run the application to access the docs$(NC)"

.PHONY: all
all: deps docs build ## Install deps, generate docs, and build
	@echo "$(GREEN)All tasks completed successfully!$(NC)"