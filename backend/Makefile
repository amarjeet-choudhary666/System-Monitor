# CodeXray Observability Service Makefile

.PHONY: run build test clean deps migrate logs

# Default target
all: deps build

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run the application
run:
	go run ./cmd/server/main.go

# Build the application
build:
	go build -o bin/codexray ./cmd/server/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f data/*.db

# Create data directory
setup:
	mkdir -p data
	mkdir -p logs

# Run with sample log analysis
demo: setup
	@echo "Starting CodeXray demo..."
	@echo "Sample log file created at: data/sample.log"
	go run ./cmd/server/main.go

# Development mode with hot reload (requires air)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		make run; \
	fi

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from: https://golangci-lint.run/usage/install/"; \
	fi

# Show help
help:
	@echo "Available targets:"
	@echo "  run     - Run the application"
	@echo "  build   - Build the application"
	@echo "  test    - Run tests"
	@echo "  clean   - Clean build artifacts"
	@echo "  deps    - Install dependencies"
	@echo "  setup   - Create necessary directories"
	@echo "  demo    - Run demo with sample data"
	@echo "  dev     - Run in development mode with hot reload"
	@echo "  fmt     - Format code"
	@echo "  lint    - Lint code"
	@echo "  help    - Show this help"