.PHONY: help build build-backend build-frontend test test-coverage run run-dev docker-build docker-run clean install-deps

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-deps: ## Install all dependencies
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing frontend dependencies..."
	cd web && npm install

build: build-backend build-frontend ## Build both backend and frontend

build-backend: ## Build the Go backend
	@echo "Building backend..."
	go build -o bin/server cmd/server/main.go
	go build -o bin/publisher cmd/publisher/main.go
	@echo "Backend built successfully!"

build-frontend: ## Build the React frontend
	@echo "Building frontend..."
	cd web && npm run build
	@echo "Frontend built successfully!"

test: ## Run all tests
	@echo "Running Go tests..."
	go test ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

run: build ## Build and run the application
	@echo "Starting server..."
	./bin/server -config config.yaml

run-dev: ## Run in development mode (backend and frontend separately)
	@echo "Starting backend server..."
	@echo "Run 'cd web && npm run dev' in another terminal for the frontend"
	go run cmd/server/main.go -config config.yaml

run-publisher: ## Run the test publisher
	@echo "Running test publisher..."
	go run cmd/publisher/main.go -stream test-stream -uri amqp://root:magic@cdvm:5672 -rate 1

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t rmq-stream-viewer:latest .
	@echo "Docker image built successfully!"

docker-run: ## Run with Docker Compose
	@echo "Starting with Docker Compose..."
	docker-compose up -d
	@echo "Application started! Access at http://localhost:8080"

docker-stop: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/dist/
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

lint: ## Run linters
	@echo "Running Go linters..."
	go vet ./...
	@echo "Running frontend linter..."
	cd web && npm run lint

format: ## Format code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Done!"

