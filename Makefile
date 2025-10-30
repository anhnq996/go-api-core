.PHONY: help build run test clean docker-build docker-up docker-down migrate seed

# Default target
help:
	@echo "Available commands:"
	@echo "  make build         - Build binary"
	@echo "  make run           - Run application"
	@echo "  make watch         - Run with hot reload (auto restart on file changes)"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "  make migrate       - Run database migrations"
	@echo "  make seed          - Run database seeders"
	@echo "  make migrate-down  - Rollback migrations"
	@echo "  make fresh         - Fresh setup (drop all + migrate + seed)"
	@echo "  make migrate-create name=your_migration - Create new migration files (up/down)"
	@echo ""
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start all services"
	@echo "  make docker-down   - Stop all services"
	@echo ""
	@echo "  make dev           - Start dev environment (postgres + redis)"
	@echo "  make setup         - Complete setup (docker + migrate + seed)"
	@echo "  make gen-keys      - Generate RSA keys to keys/private.pem & keys/public.pem"

# Build binary
build:
	@echo "Building..."
	@go build -o bin/apicore cmd/app/main.go
	@echo "✅ Build complete: bin/apicore"

# Run application
run:
	@echo "Starting application..."
	@go run cmd/app/main.go

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
watch:
	@echo "Starting application with hot reload..."
	@air

# Alias for watch
dev-watch: watch

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
	@if [ -x ./migrate ]; then \
		./migrate up; \
	else \
		go run cmd/migrate/main.go up; \
	fi

migrate-down:
	@echo "Rolling back migrations..."
	@if [ -x ./migrate ]; then \
		./migrate down; \
	else \
		go run cmd/migrate/main.go down; \
	fi

seed:
	@echo "Running seeders..."
	@if [ -x ./migrate ]; then \
		./migrate seed || go run cmd/migrate/main.go seed; \
	else \
		go run cmd/migrate/main.go seed; \
	fi

migrate-fresh:
	@echo "⚠️  Dropping all tables and re-running migrations..."
	@if [ -x ./migrate ]; then \
		./migrate fresh; \
	else \
		go run cmd/migrate/main.go fresh; \
	fi
	@echo "✅ Fresh migration complete"

migrate-version:
	@if [ -x ./migrate ]; then \
		./migrate version; \
	else \
		go run cmd/migrate/main.go version; \
	fi

# Fresh setup (drop all, migrate, seed)
fresh: migrate-fresh seed
	@echo "✅ Fresh database setup complete!"

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
	@go install github.com/air-verse/air@latest
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

# Generate RSA keys for JWT
gen-keys:
	@echo "Generating RSA keys (2048-bit) to keys/*.pem ..."
	@go run ./cmd/tools/genkeys
	@echo "✅ Keys generated"

# Migration create
migrate-create:
	@if [ -z "$(name)" ]; then \
		printf "\n❌ Vui lòng truyền biến name cho migration mới.\n\n  Ví dụ: make migrate-create name=add_products_table\n\n" && exit 1; \
	fi; \
	dir=database/migrations; \
	last=`ls $$dir | grep -E '^[0-9]{6}_.+\.up\.sql$$' | sort | tail -n 1 | sed 's/_.*//'`; \
	if [ -z "$$last" ]; then num=000001; else num=`printf "%06d" $$((10#$$last + 1))`; fi; \
	fname=$$num"_$(name)"; \
	echo "Tạo: $$dir/$$fname.up.sql & .down.sql"; \
	touch $$dir/$$fname.up.sql $$dir/$$fname.down.sql; \
	echo "-- Migration: $$fname --\n-- Viết câu lệnh SQL tại đây --" > $$dir/$$fname.up.sql; \
	echo "-- Rollback: $$fname --\n-- Viết câu lệnh rollback SQL tại đây --" > $$dir/$$fname.down.sql;

