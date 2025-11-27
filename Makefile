# Finance Tracker Backend - Makefile
# Version: 1.0 | Milestone 1

.PHONY: help run build test test-unit test-integration lint fmt clean docker-build docker-run migrate-up migrate-down tidy

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=finance-tracker-api
MAIN_PATH=./cmd/api

# Docker parameters
DOCKER_IMAGE=finance-tracker-backend
DOCKER_TAG=latest

# Default target
help: ## Show this help message
	@echo "Finance Tracker Backend - Available Commands"
	@echo "============================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# =============================================================================
# Development
# =============================================================================

run: ## Run the application
	$(GOCMD) run $(MAIN_PATH)/main.go

run-dev: ## Run with hot reload (requires air)
	air

build: ## Build the application binary
	CGO_ENABLED=0 $(GOBUILD) -ldflags="-w -s" -o bin/$(BINARY_NAME) $(MAIN_PATH)/main.go

clean: ## Remove build artifacts
	rm -rf bin/
	rm -rf tmp/

tidy: ## Tidy go modules
	$(GOMOD) tidy

# =============================================================================
# Testing
# =============================================================================

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	$(GOTEST) -v -race -cover ./internal/... ./pkg/...

test-integration: ## Run integration tests (BDD)
	$(GOTEST) -v -tags=integration ./test/integration/...

test-coverage: ## Run tests with coverage report
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# =============================================================================
# Code Quality
# =============================================================================

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	$(GOCMD) fmt ./...
	goimports -w .

vet: ## Run go vet
	$(GOCMD) vet ./...

# =============================================================================
# Database
# =============================================================================

migrate-up: ## Run database migrations up
	migrate -path scripts/migrations -database "$${DATABASE_URL}" up

migrate-down: ## Run database migrations down
	migrate -path scripts/migrations -database "$${DATABASE_URL}" down

migrate-create: ## Create a new migration (usage: make migrate-create name=create_users)
	migrate create -ext sql -dir scripts/migrations -seq $(name)

# =============================================================================
# Docker
# =============================================================================

docker-build: ## Build Docker image
	docker build -f build/docker/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-build-dev: ## Build development Docker image
	docker build -f build/docker/Dockerfile.dev -t $(DOCKER_IMAGE):dev .

# =============================================================================
# Dependencies
# =============================================================================

deps: ## Download dependencies
	$(GOMOD) download

deps-update: ## Update dependencies
	$(GOMOD) tidy
	$(GOGET) -u ./...

# =============================================================================
# Installation
# =============================================================================

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/cosmtrek/air@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
