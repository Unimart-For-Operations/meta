.PHONY: help build run test clean docker-build docker-run

.DEFAULT_GOAL := help

## Build the application binary
build:
	@echo "Building ${{ values.name }}..."
	@go build -o bin/${{ values.name }} cmd/server/main.go
	@echo "✓ Build complete: bin/${{ values.name }}"

## Run the application locally
run:
	@echo "Running ${{ values.name }}..."
	@go run cmd/server/main.go

## Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html
	@echo "✓ Clean complete"

## Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t ${{ values.name }}:latest .
	@echo "✓ Docker image built: ${{ values.name }}:latest"

## Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p ${{ values.port }}:${{ values.port }} --rm ${{ values.name }}:latest

## Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

## Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run
	@echo "✓ Linting complete"

## Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies installed"

## Show help
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'
