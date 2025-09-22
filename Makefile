.PHONY: build test clean run dev fmt lint deps

# Build configuration
BINARY_NAME=proxy
BUILD_DIR=bin
MAIN_PATH=./cmd/proxy

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

# Build for production (static binary)
build-prod:
	@echo "Building $(BINARY_NAME) for production..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built production binary $(BUILD_DIR)/$(BINARY_NAME)"

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

# Run the application with example config
run: build
	@echo "Running $(BINARY_NAME) with example config..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -config examples/config.yaml

# Run in development mode with race detection
dev:
	@echo "Running in development mode..."
	@go run -race $(MAIN_PATH) -config examples/config.yaml

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

# Generate Go modules
mod-init:
	@go mod init github.com/ubcent/edge.link

# Display help
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  build-prod  - Build for production (static binary)"
	@echo "  test        - Run tests"
	@echo "  test-race   - Run tests with race detection"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Build and run with example config"
	@echo "  dev         - Run in development mode with race detection"
	@echo "  fmt         - Format code"
	@echo "  lint        - Run linter"
	@echo "  deps        - Install/update dependencies"
	@echo "  help        - Show this help message"

# Default target
all: deps fmt test build