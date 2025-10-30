# 🚀 ApiCore - User Management API

API quản lý người dùng được xây dựng với Go, sử dụng kiến trúc modular và dependency injection.

## ⚡ Quick Reference

```bash
# Setup lần đầu
make setup            # Start services + migrate + seed
make watch            # Run app with hot reload

# Development
make dev              # Start PostgreSQL + Redis
make migrate          # Run migrations
make seed             # Run seeders
make run              # Start server (no reload)
make watch            # Start server with hot reload (auto restart)
make test             # Run tests

# Production
make docker-build     # Build Docker image
make docker-up        # Start full stack
make docker-logs      # View logs
make docker-down      # Stop all

```

## ✨ Tính Năng

- ✅ RESTful API design
- ✅ Kiến trúc modular với Chi Router
- ✅ Dependency Injection với Google Wire
- ✅ PostgreSQL database với GORM
- ✅ Redis caching với Remember pattern
- ✅ Database migrations với golang-migrate
- ✅ Swagger/OpenAPI documentation
- ✅ Interactive API documentation với Swagger UI
- ✅ Structured logging với zerolog
- ✅ Request/response logging middleware
- ✅ **Multi-language support (i18n) - EN/VI**
- ✅ **JWT Authentication & Authorization**
- ✅ **Role-based access control (RBAC)**
- ✅ **Generic Base Repository pattern**
- ✅ **FCM (Firebase Cloud Messaging) integration**
- ✅ **Hot reload với Air**
- ✅ **Utils package với 100+ helper functions**
- ✅ Health check endpoint
- ✅ Panic recovery middleware
- ✅ Request ID tracking
- ✅ Docker support (multi-stage build)
- ✅ Makefile commands

## 🏗️ Cấu Trúc Dự Án

```
ApiCore/
├── bin/                         # Build output
├── build/
│   └── docker/
│       └── entrypoint.sh
├── cmd/
│   ├── app/                     # App entry (main)
│   │   └── main.go
│   ├── migrate/                 # Migration CLI
│   │   └── main.go
│   └── tools/
│       └── genkeys/
│           └── main.go
├── config/                      # Cấu hình (go)
├── database/
│   ├── migrations/              # Migration scripts
│   └── seeders/                 # Seeder scripts
├── internal/
│   ├── app/
│   │   ├── auth/                # Module Auth
│   │   └── user/                # Module User
│   ├── models/
│   ├── repositories/
│   ├── routes/
│   ├── schedules/
│   │   └── jobs/
│   ├── templates/
│   │   └── emails/
│   └── wire/
├── keys/                        # JWT keys (.gitignore)
├── pkg/                         # Lib tái sử dụng (cache, jwt, logger...)
├── storages/                    # File lưu hoặc logs
├── test/                        # Code test
├── translations/                # Dịch thuật
├── docs/                        # Tài liệu, swagger
├── Dockerfile*
├── Makefile
└── README.md
```

## 🚀 Quick Start

### Yêu Cầu

- Go 1.23.4 hoặc cao hơn
- PostgreSQL 15+ (hoặc Docker)
- Wire CLI (cho dependency injection)

### Cài Đặt

1. **Clone repository**

    ```bash
    git clone <repository-url>
    cd ApiCore
    ```

2. **Install dependencies**
    ```bash
    go mod download
    ```

3. **(Optional, if not available) Install Wire CLI**
    ```bash
    go install github.com/google/wire/cmd/wire@latest
    ```

4. **Prepare environment config**
    ```bash
    cp env.example .env
    # Edit .env for your database/Redis as needed
    ```

5. **Generate RSA keys for JWT (one time only)**
    ```bash
    make gen-keys
    # Output: keys/private.pem & keys/public.pem
    ```

6. **Start infrastructure (PostgreSQL + Redis)**
    ```bash
    make dev
    # Wait a few seconds for services to be ready
    ```

7. **Run migrations and seed database**
    ```bash
    make migrate
    make seed
    ```

8. **Run the application**
    ```bash
    make run
    # or for hot reload (recommended during dev):
    make watch
    ```

**Note:** All important commands are defined in the Makefile for easy usage during both development and production.

---

## 📚 Documentation

Truy cập documentation tại: **http://localhost:3000/docs**

### Các Trang Documentation

| URL             | Mô Tả                                |
| --------------- | ------------------------------------ |
| `/docs`         | Trang chủ documentation với overview |
| `/swagger`      | Swagger UI - Interactive API testing |
| `/swagger.json` | OpenAPI specification file           |

### Hướng dẫn chi tiết

- [**Development Guide**](docs/development-guide.md) - Hướng dẫn phát triển

### Package Documentation

- [**pkg/jwt**](pkg/jwt/README.md) - JWT authentication & authorization 🌟
- [**pkg/validator**](pkg/validator/README.md) - Auto validation với struct tags 🌟
- [**pkg/response**](pkg/response/README.md) - Standardized REST API response 🌟
- [**pkg/i18n**](pkg/i18n/README.md) - Internationalization (i18n) support 🌟
- [**pkg/utils**](pkg/utils/README.md) - Common utility functions 🌟
- [**pkg/fcm**](pkg/fcm/README.md) - Firebase Cloud Messaging 🌟
- [pkg/logger](pkg/logger/README.md) - Structured logging
- [pkg/cache](pkg/cache/README.md) - Redis caching utilities
- [internal/repositories](internal/repositories/README.md) - Generic Base Repository pattern 🌟

## 🛣️ API Endpoints

### Health Check

- `GET /ping` - Kiểm tra server status

### User Management

- `GET /api/v1/users` - Lấy danh sách users
- `POST /api/v1/users` - Tạo user mới
- `GET /api/v1/users/{id}` - Lấy user theo ID
- `PUT /api/v1/users/{id}` - Cập nhật user
- `DELETE /api/v1/users/{id}` - Xóa user

Chi tiết xem tại [Swagger UI](http://localhost:3000/swagger)

## 🏗️ Kiến Trúc

### Module-Based Architecture

Mỗi module (user, order, product...) được tổ chức theo cấu trúc:

```
module/
├── controller.go   # HTTP handlers
├── service.go      # Business logic
└── route.go        # Routes definition
```

### Dependency Injection với Wire

Wire tự động generate code để inject dependencies:

- Repository → Service → Handler → Router

Không cần khởi tạo thủ công từng dependency trong `main.go`.

## 🔧 Thêm Module Mới

### Bước 1: Tạo cấu trúc module

```bash
mkdir -p internal/app/order
```

### Bước 2: Tạo các file cần thiết

**controller.go**

```go
package order

type Handler struct {
    service *Service
}

func NewHandler(svc *Service) *Handler {
    return &Handler{service: svc}
}

// Implement handlers: Index, Show, Store, Update, Destroy
```

**service.go**

```go
package order

type Service struct {
    repo repository.OrderRepository
}

func NewService(r repository.OrderRepository) *Service {
    return &Service{repo: r}
}

// Implement business logic
```

**route.go**

```go
package order

func RegisterRoutes(r chi.Router, h *Handler) {
    r.Route("/orders", func(r chi.Router) {
        r.Get("/", h.Index)
        r.Post("/", h.Store)
        // ...
    })
}
```

### Bước 3: Cập nhật Wire configuration

**internal/wire/wire.go**

```go
func InitializeApp() *routes.Controllers {
    wire.Build(
        // Repositories
        repository.NewUserRepository,
        repository.NewOrderRepository,  // Thêm

        // Services
        user.NewService,
        order.NewService,  // Thêm

        // Handlers
        user.NewHandler,
        order.NewHandler,  // Thêm

        // Controllers
        routes.NewControllers,
    )
    return nil
}
```

### Bước 4: Cập nhật routes

**internal/routes/routes.go**

```go
type Controllers struct {
    UserHandler  *user.Handler
    OrderHandler *order.Handler  // Thêm
}

func NewControllers(
    userHandler *user.Handler,
    orderHandler *order.Handler,  // Thêm
) *Controllers {
    return &Controllers{
        UserHandler:  userHandler,
        OrderHandler: orderHandler,  // Thêm
    }
}

func RegisterRoutes(r chi.Router, c *Controllers) {
    r.Route("/api/v1", func(r chi.Router) {
        user.RegisterRoutes(r, c.UserHandler)
        order.RegisterRoutes(r, c.OrderHandler)  // Thêm
    })
}
```

### Bước 5: Generate Wire code

```bash
wire ./internal/wire
```

### Bước 6: Cập nhật Swagger documentation

Thêm endpoints mới vào `docs/swagger.json`

## 🛠️ Makefile Commands

### Development Commands

```bash
make dev              # Start dev services (PostgreSQL + Redis)
make dev-down         # Stop dev services
make run              # Run application locally
make build            # Build binary to bin/apicore
make clean            # Clean build artifacts
make test             # Run tests
make fmt              # Format code
make tidy             # Tidy dependencies
```

### Database Commands

```bash
make migrate          # Run database migrations
make migrate-down     # Rollback all migrations
make migrate-version  # Show migration version
make seed             # Run database seeders
make setup            # Complete setup (dev + migrate + seed)
```

### Docker Commands

```bash
make docker-build     # Build Docker image
make docker-up        # Start all services (production)
make docker-down      # Stop all services
make docker-logs      # View logs from all services
```

### Utility Commands

```bash
make wire             # Generate Wire DI code
make install-tools    # Install dev tools (wire, migrate)
make lint             # Run linter
make check            # Run all checks (fmt + lint + test)
```

## 🐳 Docker Commands

### Build Image

```bash
# Using Make
make docker-build

# Direct Docker command
docker build -t apicore:latest .

# With specific version
docker build -t apicore:v1.0.0 .
```

### Docker Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker logs -f apicore-api
docker logs -f apicore-postgres
docker logs -f apicore-redis

# Last 100 lines
docker logs --tail 100 apicore-api
```

## 📦 Dependencies

- [chi](https://github.com/go-chi/chi) - Lightweight router
- [wire](https://github.com/google/wire) - Dependency injection
- [gorm](https://gorm.io) - ORM library
- [go-redis](https://github.com/redis/go-redis) - Redis client
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations
- [uuid](https://github.com/google/uuid) - UUID generation
- [zerolog](https://github.com/rs/zerolog) - Structured logging
- PostgreSQL 15+ - Database
- Redis 7+ - Cache

## 🔒 Security

- Request ID tracking
- Panic recovery
- Input validation
- Rate limiting
- Authentication

## 🚀 Deployment

## 📋 All Commands Cheat Sheet

### Quick Start (One-Time Setup)

```bash
make setup            # Start services + migrate + seed (all-in-one)
make run              # Start application
```

### Development Workflow

```bash
# Start infrastructure
make dev              # Start PostgreSQL + Redis
make migrate          # Run migrations
make seed             # Seed database

# Run application
make run              # Start server
# or
go run cmd/app/main.go

# Test
curl http://localhost:3000/api/v1/users
```

### Build & Deploy

```bash
# Local build
make build            # Output: bin/apicore
./bin/apicore         # Run binary

# Docker build
make docker-build     # Build image
docker images apicore # Check size (~20-30MB)

# Docker run (development)
docker-compose up -d  # Infrastructure only

# Docker run (production)
make docker-up        # Start full stack (API + infra)
make docker-logs      # View logs
make docker-down      # Stop all
```

### Cache Operations

```bash
# Check Redis
docker exec -it apicore-redis redis-cli

# Inside redis-cli
KEYS *                # List all keys
GET users:all         # Get cached users
TTL users:all         # Check TTL
DEL users:all         # Delete cache
FLUSHDB               # Clear all cache
```

### Testing

```bash
# Run tests
make test
# or
go test -v ./...

# Run specific test
go test -v ./internal/app/user/...

# With coverage
go test -v -cover ./...
```

### Code Quality

```bash
make fmt              # Format code
make tidy             # Tidy dependencies
make lint             # Run linter (if installed)
make check            # Run all checks
```

### Wire (Dependency Injection)

```bash
make wire             # Generate Wire code
# or
wire ./internal/wire
```

### Utilities

```bash
make clean            # Clean build artifacts
make install-tools    # Install wire, migrate CLI
make help             # Show available commands
```

## 👥 Authors

- AnhNQ
