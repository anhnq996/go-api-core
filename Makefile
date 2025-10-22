.PHONY: help build run test clean docker-build docker-up docker-down migrate seed

# Default target
help:
	@echo "Available commands:"
	@echo "  make build         - Build binary"
	@echo "  make run           - Run application"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "  make migrate       - Run database migrations"
	@echo "  make seed          - Run database seeders"
	@echo "  make migrate-down  - Rollback migrations"
	@echo ""
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start all services"
	@echo "  make docker-down   - Stop all services"
	@echo ""
	@echo "  make dev           - Start dev environment (postgres + redis)"
	@echo "  make setup         - Complete setup (docker + migrate + seed)"

# Build binary
build:
	@echo "Building..."
	@go build -o bin/apicore cmd/app/main.go
	@echo "✅ Build complete: bin/apicore"

# Run application
run:
	@echo "Starting application..."
	@go run cmd/app/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f apicore
	@echo "✅ Clean complete"

# Database commands
migrate:
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back migrations..."
	@go run cmd/migrate/main.go down

seed:
	@echo "Running seeders..."
	@go run cmd/migrate/main.go seed

migrate-version:
	@go run cmd/migrate/main.go version

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t apicore:latest .
	@echo "✅ Docker image built"

docker-up:
	@echo "Starting services..."
	@docker-compose -f docker-compose.prod.yml up -d
	@echo "✅ Services started"

docker-down:
	@echo "Stopping services..."
	@docker-compose -f docker-compose.prod.yml down
	@echo "✅ Services stopped"

docker-logs:
	@docker-compose -f docker-compose.prod.yml logs -f

# Development commands
dev:
	@echo "Starting dev environment..."
	@docker-compose up -d postgres redis
	@echo "✅ Dev environment ready"

dev-down:
	@docker-compose down

# Complete setup
setup: dev
	@echo "Waiting for services to be ready..."
	@sleep 3
	@make migrate
	@make seed
	@echo "✅ Setup complete! Run 'make run' to start the app"

# Wire generate
wire:
	@echo "Generating Wire code..."
	@wire ./internal/wire
	@echo "✅ Wire code generated"

# Install tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/google/wire/cmd/wire@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "✅ Tools installed"

# Lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "✅ Dependencies tidied"

# Full check (before commit)
check: fmt lint test
	@echo "✅ All checks passed"

