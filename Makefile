.PHONY: all build test clean run lint vet fmt generate generate-swagger test-coverage test-integration deps

# Variables
BINARY_NAME=erp-system
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GO?=go
GOFMT=go fmt
GOMOD=$(GO) mod
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOLINT=golangci-lint
GOVET=$(GO) vet

# Directories
ROOT_DIR=$(shell pwd)
BUILD_DIR=$(ROOT_DIR)/build
BIN_DIR=$(ROOT_DIR)/bin

# Docker
DOCKER=$(shell which docker)
DOCKER_COMPOSE=$(shell which docker-compose)

# Test coverage
COVERAGE_DIR=$(ROOT_DIR)/coverage

# Default target
all: deps build test

# Download dependencies
deps:
	@echo "Downloading Go dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Build all services
build: deps
	@echo "Building all services..."
	$(GOBUILD) -o $(BIN_DIR)/client-command-service ./cmd/client-command-service
	$(GOBUILD) -o $(BIN_DIR)/auth-service ./cmd/auth-service
	$(GOBUILD) -o $(BIN_DIR)/client-query-service ./cmd/client-query-service
	$(GOBUILD) -o $(BIN_DIR)/api-gateway ./cmd/api-gateway
	@echo "Build complete. Binaries in $(BIN_DIR)/"

# Build specific service
build/%:
	$(GOBUILD) -o $(BIN_DIR)/$* ./cmd/$*

# Run all tests
test: deps
	@echo "Running all tests..."
	$(GOTEST) -v ./...
	@echo "All tests passed!"

# Run tests with coverage
test-coverage: deps
	@echo "Running tests with coverage..."
	mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	$(GOTEST) -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@echo "Coverage report generated in $(COVERAGE_DIR)/"

# Run integration tests
test-integration: deps
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

# Run linter
lint: deps
	@echo "Running linter..."
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run ./...; \
	else \
		echo "golangci-lint not installed. Run: brew install golangci-lint"; \
	fi

# Run go vet
vet: deps
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) ./...

# Generate code
generate:
	@echo "Running code generators..."
	$(GOFMT) -s ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/*
	rm -rf $(BUILD_DIR)/*
	rm -rf $(COVERAGE_DIR)/*
	@echo "Clean complete!"

# Run the client-command-service
run-client-command:
	@echo "Starting client-command-service..."
	$(BIN_DIR)/client-command-service

# Run the auth-service
run-auth:
	@echo "Starting auth-service..."
	$(BIN_DIR)/auth-service

# Run the client-query-service
run-client-query:
	@echo "Starting client-query-service..."
	$(BIN_DIR)/client-query-service

# Run the api-gateway
run-api-gateway:
	@echo "Starting api-gateway..."
	$(BIN_DIR)/api-gateway

# Run all services (requires docker-compose)
run-services:
	@echo "Starting all services..."
	@if command -v $(DOCKER_COMPOSE) >/dev/null 2>&1; then \
		$(DOCKER_COMPOSE) up -d; \
	else \
		echo "docker-compose not installed."; \
	fi

# Stop all services
stop-services:
	@echo "Stopping all services..."
	@if command -v $(DOCKER_COMPOSE) >/dev/null 2>&1; then \
		$(DOCKER_COMPOSE) down; \
	else \
		echo "docker-compose not installed."; \
	fi

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	@for service in client-command-service auth-service client-query-service api-gateway; do \
		docker build -t ims-erp/$$service:$(VERSION) -t ims-erp/$$service:latest ./cmd/$$service; \
	done
	@echo "Docker images built!"

# Push Docker images
docker-push:
	@echo "Pushing Docker images..."
	@for service in client-command-service auth-service client-query-service api-gateway; do \
		docker push ims-erp/$$service:$(VERSION); \
	done
	@echo "Docker images pushed!"

# Show help
help:
	@echo "ERP System Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all              - Download deps, build, and test (default)"
	@echo "  deps             - Download Go dependencies"
	@echo "  build            - Build all services"
	@echo "  build/<service>  - Build specific service"
	@echo "  test             - Run all tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  test-integration - Run integration tests"
	@echo "  lint             - Run linter"
	@echo "  vet              - Run go vet"
	@echo "  fmt              - Format Go code"
	@echo "  generate         - Run code generators"
	@echo "  clean            - Clean build artifacts"
	@echo "  run-client-command - Run client-command-service"
	@echo "  run-auth         - Run auth-service"
	@echo "  run-client-query - Run client-query-service"
	@echo "  run-api-gateway  - Run api-gateway"
	@echo "  run-services     - Start all services (docker-compose)"
	@echo "  stop-services    - Stop all services"
	@echo "  docker-build     - Build Docker images"
	@echo "  docker-push      - Push Docker images"
	@echo "  help             - Show this help"

# Version info
version:
	@echo "ERP System Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Git Commit: $(GIT_COMMIT)"
