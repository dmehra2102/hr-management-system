# HR Management System Makefile

.PHONY: help build run test clean proto migrate-up migrate-down docker-build docker-run

# Default target
help:
	@echo "Available commands:"
	@echo "  build          Build the application"
	@echo "  run            Run the application locally"
	@echo "  test           Run all tests"
	@echo "  test-unit      Run unit tests only"
	@echo "  test-integration Run integration tests only"
	@echo "  clean          Clean build artifacts"
	@echo "  proto          Generate protobuf files"
	@echo "  migrate-up     Run database migrations"
	@echo "  migrate-down   Rollback database migrations"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run with Docker Compose"
	@echo "  docker-down    Stop Docker Compose"
	@echo "  lint           Run linter"
	@echo "  fmt            Format code"

# Build the application
build:
	@echo "Building HR Management System..."
	go build -ldflags="-w -s" -o bin/hr-server ./cmd/server

# Run the application locally
run:
	@echo "Starting HR Management System..."
	go run ./cmd/server

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v -run Integration ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf logs/
	go clean

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	protoc --proto_path=api\proto\v1 --go_out=. --go-grpc_out=. .\api\proto\v1\auth.proto .\api\proto\v1\department.proto .\api\proto\v1\employee.proto .\api\proto\v1\leave.proto .\api\proto\v1\performance.proto

# Run database migrations up
migrate-up:
	@echo "Running database migrations..."
	@chmod +x scripts/migrate-up.sh
	./scripts/migrate-up.sh

# Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	@chmod +x scripts/migrate-down.sh
	./scripts/migrate-down.sh

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t hr-management-system:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose
docker-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Install development dependencies
install-tools:
	@echo "Installing development tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Initialize the project
init: install-tools
	@echo "Initializing project..."
	go mod tidy
	make proto
	make migrate-up

# Run in development mode with auto-reload
dev:
	@echo "Running in development mode..."
	air -c .air.toml
