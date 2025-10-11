# CodeXray Observability Service Makefile

.PHONY: run build test clean deps

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


fmt:
	go fmt ./...

# Show help
help:
	@echo "Available targets:"
	@echo "  run     - Run the application"
	@echo "  build   - Build the application"
	@echo "  test    - Run tests"
	@echo "  clean   - Clean build artifacts"
	@echo "  deps    - Install dependencies"
	@echo "  fmt     - Format code"
	@echo "  help    - Show this help"