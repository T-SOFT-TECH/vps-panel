.PHONY: help dev build start stop clean install backend frontend

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies for both frontend and backend
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Installing backend dependencies..."
	cd backend && go mod download
	@echo "✅ Dependencies installed!"

dev: ## Run development servers (frontend + backend)
	@echo "Starting development servers..."
	@make -j2 dev-frontend dev-backend

dev-frontend: ## Run frontend development server
	cd frontend && npm run dev

dev-backend: ## Run backend development server
	cd backend && go run cmd/server/main.go

build: ## Build both frontend and backend for production
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Building backend..."
	cd backend && go build -o server cmd/server/main.go
	@echo "✅ Build complete!"

docker-build: ## Build Docker images
	docker-compose build

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

clean: ## Clean build artifacts and dependencies
	rm -rf frontend/node_modules
	rm -rf frontend/build
	rm -rf frontend/.svelte-kit
	rm -rf backend/server
	rm -rf backend/data
	@echo "✅ Cleaned!"

test-backend: ## Run backend tests
	cd backend && go test ./...

format-backend: ## Format backend code
	cd backend && go fmt ./...

format-frontend: ## Format frontend code
	cd frontend && npm run format

lint-frontend: ## Lint frontend code
	cd frontend && npm run lint
