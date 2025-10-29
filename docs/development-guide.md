# Development Guide

Hướng dẫn phát triển dự án ApiCore.

## Prerequisites

- Go 1.23.4+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (optional)
- Make
- Air (for hot reload)

## Quick Start

### 1. Clone & Setup

```bash
# Clone repository
git clone <repository-url>
cd ApiCore

# Install development tools
make install-tools

# Setup environment
cp env.example .env
```

### 2. Start Development Environment

```bash
# Terminal 1: Start services (PostgreSQL + Redis)
make dev

# Terminal 2: Run migrations & seeders
make migrate
make seed

# Terminal 3: Start app with hot reload
make watch
```

Truy cập: http://localhost:3000

## Development Workflow

### Option 1: Hot Reload (Recommended)

```bash
# Start với auto-restart khi code thay đổi
make watch

# Thay đổi code trong internal/app/user/controller.go
# -> Air tự động rebuild & restart
# -> Kiểm tra http://localhost:3000
```

### Option 2: Manual Restart

```bash
# Start server
make run

# Khi thay đổi code, Ctrl+C và chạy lại
make run
```

## Project Structure

```
ApiCore/
├── cmd/app/main.go          # Application entry point
├── internal/                # Private application code
│   ├── app/                 # Application modules
│   │   └── user/           # User module
│   ├── models/             # Data models
│   ├── repositories/       # Data access layer
│   └── routes/             # Route registration
├── pkg/                    # Public packages
│   ├── cache/              # Redis cache
│   ├── fcm/                # Firebase Cloud Messaging
│   ├── i18n/               # Internationalization
│   ├── logger/             # Logging
│   └── response/           # API response format
├── config/                 # Configuration
├── database/               # Migrations & seeders
├── translations/           # i18n translation files
└── docs/                   # Documentation
```

## Making Changes

### 1. Create New Module

```bash
# Tạo thư mục module mới
mkdir -p internal/app/product

# Tạo files
touch internal/app/product/controller.go
touch internal/app/product/service.go
touch internal/app/product/route.go
```

### 2. Update Wire DI

```go
// internal/wire/wire.go
func InitializeApp(db *gorm.DB, cache cache.Cache) *Controllers {
    wire.Build(
        // ... existing providers
        product.NewRepository,
        product.NewService,
        product.NewHandler,
    )
    return &Controllers{}
}
```

```bash
# Generate wire code
make wire
```

### 3. Register Routes

```go
// internal/routes/routes.go
func RegisterRoutes(r *chi.Mux, controllers *wire.Controllers) {
    // ... existing routes

    // Product routes
    product.RegisterRoutes(r, controllers.ProductHandler)
}
```

### 4. Test Changes

```bash
# Hot reload sẽ tự động restart
# Hoặc chạy manual
make run
```

## Database Migrations

### Create Migration

```bash
# Create new migration
migrate create -ext sql -dir database/migrations -seq create_products_table

# Edit files:
# database/migrations/000002_create_products_table.up.sql
# database/migrations/000002_create_products_table.down.sql
```

### Run Migrations

```bash
# Run all pending migrations
make migrate

# Rollback last migration
make migrate-down

# Check migration version
make migrate-version
```

## Testing

### Run Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/app/user/...

# Run with coverage
go test -cover ./...

# Run with verbose
go test -v ./...
```

### Write Tests

```go
// internal/app/user/service_test.go
package user_test

import (
    "testing"
    "api-core/internal/app/user"
)

func TestGetUser(t *testing.T) {
    // Setup
    service := user.NewService(mockRepo)

    // Execute
    user, err := service.GetByID("123")

    // Assert
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if user.ID != "123" {
        t.Errorf("Expected user ID 123, got %s", user.ID)
    }
}
```

## Code Quality

### Format Code

```bash
# Format all code
make fmt

# Check formatting
gofmt -l .
```

### Lint Code

```bash
# Run linter
make lint

# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Tidy Dependencies

```bash
# Tidy go.mod
make tidy
```

### Run All Checks

```bash
# Format + Lint + Test
make check
```

## Debugging

### 1. Enable Debug Logging

```go
// cmd/app/main.go
logger.Init(logger.Config{
    Level: "debug",  // Changed from "info"
    // ...
})
```

### 2. View Logs

```bash
# Application logs
tail -f storages/logs/app.log

# Build errors (when using Air)
tail -f build-errors.log

# Docker logs
make docker-logs
```

### 3. Debug with Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Run with debugger
dlv debug ./cmd/app/main.go

# In delve console
(dlv) break main.main
(dlv) continue
```

## Environment Variables

### Development (.env)

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=apicore

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Server
SERVER_PORT=3000
```

### Production

```bash
# Set environment variables
export DB_HOST=production-db-host
export DB_PASSWORD=secure-password
export SERVER_PORT=8080

# Run
./bin/apicore
```

## Git Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/add-product-module
```

### 2. Make Changes

```bash
# Edit code
# Test changes with: make watch

# Stage changes
git add .

# Commit
git commit -m "feat: add product module"
```

### 3. Push & PR

```bash
git push origin feature/add-product-module

# Create Pull Request on GitHub/GitLab
```

## Common Tasks

### Add New Response Code

```go
// 1. Add to pkg/response/codes.go
const CodeProductNotFound = "PRODUCT_NOT_FOUND"

// 2. Add to translations/en.json
"PRODUCT_NOT_FOUND": "Product not found"

// 3. Add to translations/vi.json
"PRODUCT_NOT_FOUND": "Không tìm thấy sản phẩm"

// 4. Use in controller
response.NotFound(w, lang, response.CodeProductNotFound)
```

### Add New Cache Key

```go
// pkg/cache/keys.go (create if needed)
const (
    CacheKeyUsers    = "users:all"
    CacheKeyUser     = "users:%s"
    CacheKeyProducts = "products:all"
)

// Use in service
cache.Remember("products:all", 5*time.Minute, func() (interface{}, error) {
    return repo.GetAll()
})
```

### Add New Logger Field

```go
logger.WithFields(map[string]interface{}{
    "user_id":  userID,
    "action":   "create_product",
    "product_id": productID,
}).Info("Product created")
```

## Performance Optimization

### 1. Database Queries

```go
// Bad: N+1 query
for _, user := range users {
    orders := repo.GetOrdersByUserID(user.ID)
}

// Good: Preload
users := repo.GetUsersWithOrders()
```

### 2. Caching

```go
// Cache frequently accessed data
cache.Remember("popular-products", 5*time.Minute, func() (interface{}, error) {
    return repo.GetPopularProducts()
})
```

### 3. Pagination

```go
// Always paginate large datasets
func ListProducts(w http.ResponseWriter, r *http.Request) {
    page := getPageFromRequest(r)
    perPage := 20

    products, total := service.GetPaginated(page, perPage)
    meta := response.PaginationFromRequest(r, total)

    response.SuccessWithMeta(w, lang, response.CodeSuccess, products, meta)
}
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -ti:3000

# Kill process
kill -9 <pid>
```

### Database Connection Error

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U postgres -d apicore

# Restart PostgreSQL
docker restart apicore-postgres
```

### Redis Connection Error

```bash
# Check Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost ping

# Restart Redis
docker restart apicore-redis
```

### Wire Generation Error

```bash
# Clean wire cache
rm internal/wire/wire_gen.go

# Regenerate
make wire
```

## Resources

- [Go Documentation](https://go.dev/doc/)
- [Chi Router](https://github.com/go-chi/chi)
- [GORM](https://gorm.io/docs/)
- [Air](https://github.com/air-verse/air)
- [Wire](https://github.com/google/wire)
- [Project README](../README.md)
- [Air Setup Guide](./air-setup-guide.md)
- [Response & I18n Guide](./response-and-i18n-guide.md)
