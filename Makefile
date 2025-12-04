.PHONY: help build run test clean docker-up docker-down docker-build seed

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go application
	go build -o bin/api cmd/api/main.go
	go build -o bin/seed cmd/seed/main.go

run: ## Run the application locally
	go run cmd/api/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/

docker-build: ## Build Docker images
	docker compose build

docker-up: ## Start all services with Docker Compose
	docker compose up -d

docker-down: ## Stop all Docker services
	docker compose down

docker-logs: ## View Docker logs
	docker compose logs -f api

docker-reset: ## Reset Docker environment (removes volumes)
	docker compose down -v
	docker compose up --build -d

seed: ## Run database seeder
	go run cmd/seed/main.go

tidy: ## Tidy Go modules
	go mod tidy

deps: ## Download dependencies
	go mod download
