.PHONY: build test clean run dev fmt lint deps build-admin run-admin dev-admin ui-install ui-dev ui-build docker-build docker-run help

# Build configuration
BINARY_NAME=proxy
ADMIN_BINARY_NAME=admin-api
BUILD_DIR=bin
MAIN_PATH=./cmd/proxy
ADMIN_PATH=./cmd/admin-api

# Build the proxy application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Build the admin API
build-admin:
	@echo "Building $(ADMIN_BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(ADMIN_BINARY_NAME) $(ADMIN_PATH)
	@echo "Built $(BUILD_DIR)/$(ADMIN_BINARY_NAME)"

# Build all
build-all: build build-admin

# Build for production (static binary)
build-prod:
	@echo "Building $(BINARY_NAME) for production..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built production binary $(BUILD_DIR)/$(BINARY_NAME)"

# Build admin API for production
build-admin-prod:
	@echo "Building $(ADMIN_BINARY_NAME) for production..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(ADMIN_BINARY_NAME) $(ADMIN_PATH)
	@echo "Built production binary $(BUILD_DIR)/$(ADMIN_BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	@go test -race -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean

# Run the proxy application with example config
run: build
	@echo "Running $(BINARY_NAME) with example config..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -config examples/config.yaml

# Run the admin API
run-admin: build-admin
	@echo "Running $(ADMIN_BINARY_NAME)..."
	@./$(BUILD_DIR)/$(ADMIN_BINARY_NAME) -addr :3001

# Run proxy in development mode with race detection
dev:
	@echo "Running proxy in development mode..."
	@go run -race $(MAIN_PATH) -config examples/config.yaml

# Run admin API in development mode
dev-admin:
	@echo "Running admin API in development mode..."
	@go run -race $(ADMIN_PATH) -addr :3001

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run

# Install/update dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Admin UI targets
ui-install:
	@echo "Installing admin UI dependencies..."
	@cd admin-ui && npm install

ui-dev:
	@echo "Starting admin UI in development mode..."
	@cd admin-ui && npm run dev

ui-build:
	@echo "Building admin UI for production..."
	@cd admin-ui && npm run build

# Docker targets
docker-build:
	@echo "Building Docker images..."
	@docker build -t edge-proxy .
	@docker build -f Dockerfile.admin-api -t admin-api .

docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose up

docker-stop:
	@echo "Stopping Docker services..."
	@docker-compose down

# Generate Go modules
mod-init:
	@go mod init github.com/ubcent/edge.link

# Display help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build            - Build the proxy service"
	@echo "  build-admin      - Build the admin API"
	@echo "  build-all        - Build all services"
	@echo "  build-prod       - Build proxy for production"
	@echo "  build-admin-prod - Build admin API for production"
	@echo ""
	@echo "Run targets:"
	@echo "  run              - Build and run proxy with example config"
	@echo "  run-admin        - Build and run admin API"
	@echo "  dev              - Run proxy in development mode"
	@echo "  dev-admin        - Run admin API in development mode"
	@echo ""
	@echo "Test targets:"
	@echo "  test             - Run tests"
	@echo "  test-race        - Run tests with race detection"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo ""
	@echo "UI targets:"
	@echo "  ui-install       - Install admin UI dependencies"
	@echo "  ui-dev           - Start admin UI development server"
	@echo "  ui-build         - Build admin UI for production"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build     - Build Docker images"
	@echo "  docker-run       - Start with Docker Compose"
	@echo "  docker-stop      - Stop Docker services"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean            - Clean build artifacts"
	@echo "  fmt              - Format code"
	@echo "  lint             - Run linter"
	@echo "  deps             - Install/update dependencies"
	@echo "  help             - Show this help message"

# Default target
all: deps fmt test build-all