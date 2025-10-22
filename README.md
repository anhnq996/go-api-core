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

# Database
go run cmd/migrate/main.go up       # Migrations
go run cmd/migrate/main.go seed     # Seeders
go run cmd/migrate/main.go version  # Check version
```

## ✨ Tính Năng

- ✅ RESTful API design
- ✅ Kiến trúc modular với Chi Router
- ✅ Dependency Injection với Google Wire
- ✅ PostgreSQL database với GORM
- ✅ Redis caching với Remember pattern
- ✅ Database migrations với golang-migrate
- ✅ Distributed locking
- ✅ Swagger/OpenAPI documentation
- ✅ Interactive API documentation với Swagger UI
- ✅ Structured logging với zerolog
- ✅ Request/response logging middleware
- ✅ **Standardized REST API Response format**
- ✅ **Multi-language support (i18n) - EN/VI**
- ✅ **JWT Authentication & Authorization**
- ✅ **Role-based access control (RBAC)**
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
├── cmd/
│   └── app/
│       └── main.go                 # Entry point
├── internal/
│   ├── app/
│   │   └── user/                   # User module
│   │       ├── controller.go       # HTTP handlers
│   │       ├── service.go          # Business logic
│   │       └── route.go            # Routes definition
│   ├── models/
│   │   └── user.go                 # Data models
│   ├── repositories/
│   │   ├── repository.go           # Base repository
│   │   └── user_repository.go      # User repository
│   ├── routes/
│   │   └── routes.go               # Routes registry
│   └── wire/
│       ├── wire.go                 # Wire configuration
│       └── wire_gen.go             # Generated DI code
├── pkg/
│   ├── cache/                      # Redis cache utilities
│   ├── fcm/                        # Firebase Cloud Messaging
│   ├── i18n/                       # Internationalization (EN/VI)
│   ├── jwt/                        # JWT authentication
│   ├── logger/                     # Structured logging
│   ├── response/                   # Standardized REST API response
│   └── utils/                      # Common helper functions
├── translations/
│   ├── en.json                     # English translations
│   └── vi.json                     # Vietnamese translations
├── docs/
│   ├── index.html                  # Documentation home
│   ├── swagger.html                # Swagger UI
│   ├── swagger.json                # OpenAPI specification
│   ├── routes.md                   # Routes guide
│   ├── swagger-guide.md            # Swagger usage guide
│   └── response-and-i18n-guide.md  # Response & I18n guide
├── go.mod
└── go.sum
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

2. **Cài đặt dependencies**

```bash
go mod download
```

3. **Cài đặt Wire CLI** (nếu chưa có)

```bash
go install github.com/google/wire/cmd/wire@latest
```

4. **Setup environment**

```bash
cp env.example .env
# Điều chỉnh database config trong .env
```

5. **Start PostgreSQL**

```bash
docker-compose up -d postgres
```

6. **Run migrations**

```bash
go run cmd/migrate/main.go up
```

7. **Run seeders** (optional - tạo dữ liệu mẫu)

```bash
go run cmd/migrate/main.go seed
```

8. **Start server**

```bash
go run cmd/app/main.go
```

Server sẽ khởi động tại `http://localhost:3000`

### Test API

```bash
# Health check
curl http://localhost:3000/ping

# Lấy danh sách users
curl http://localhost:3000/api/v1/users

# Tạo user mới
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Nguyễn Văn A","email":"nguyenvana@example.com"}'

# Lấy user theo ID
curl http://localhost:3000/api/v1/users/{id}

# Cập nhật user
curl -X PUT http://localhost:3000/api/v1/users/{id} \
  -H "Content-Type: application/json" \
  -d '{"name":"Nguyễn Văn B","email":"nguyenvanb@example.com"}'

# Xóa user
curl -X DELETE http://localhost:3000/api/v1/users/{id}
```

## 📚 Documentation

Truy cập documentation tại: **http://localhost:3000/docs**

### Các Trang Documentation

| URL             | Mô Tả                                |
| --------------- | ------------------------------------ |
| `/docs`         | Trang chủ documentation với overview |
| `/swagger`      | Swagger UI - Interactive API testing |
| `/swagger.json` | OpenAPI specification file           |

### Hướng dẫn chi tiết

- [**JWT Authentication Guide**](docs/jwt-guide.md) - Hướng dẫn JWT authentication 🌟
- [**Development Guide**](docs/development-guide.md) - Hướng dẫn phát triển
- [Routes Documentation](docs/routes.md) - Chi tiết về các API endpoints
- [Swagger Guide](docs/swagger-guide.md) - Hướng dẫn sử dụng Swagger
- [Docker Setup](DOCKER.md) - Hướng dẫn Docker
- [Loki + Grafana Setup](docs/loki-grafana-setup.md) - Hướng dẫn setup logging

### Package Documentation

- [**pkg/jwt**](pkg/jwt/README.md) - JWT authentication & authorization 🌟
- [**pkg/response**](pkg/response/README.md) - Standardized REST API response 🌟
- [**pkg/i18n**](pkg/i18n/README.md) - Internationalization (i18n) support 🌟
- [**pkg/utils**](pkg/utils/README.md) - Common utility functions 🌟
- [**pkg/fcm**](pkg/fcm/README.md) - Firebase Cloud Messaging 🌟
- [pkg/logger](pkg/logger/README.md) - Structured logging
- [pkg/cache](pkg/cache/README.md) - Redis caching utilities

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

### Repository Pattern

Sử dụng in-memory repository, dễ dàng chuyển sang database:

```go
// Hiện tại: In-memory
userRepo := repository.NewUserRepository()

// Tương lai: Database
userRepo := repository.NewUserRepository(db)
```

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

Chi tiết xem tại [docs/routes.md](docs/routes.md)

## 🧪 Testing

### Manual Testing với Swagger UI

1. Truy cập http://localhost:3000/swagger
2. Chọn endpoint muốn test
3. Click "Try it out"
4. Nhập parameters/body
5. Click "Execute"

### Testing với curl

Xem phần "Test API" ở trên

### Testing với Postman

Import file `swagger.json` vào Postman:

1. Mở Postman
2. Import > Link > `http://localhost:3000/swagger.json`
3. Test các endpoints

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

### Run with Docker Compose

```bash
# Development (only infrastructure)
docker-compose up -d postgres redis

# Production (full stack with API)
docker-compose -f docker-compose.prod.yml up -d

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

### Run Migrations in Docker

```bash
# Method 1: Using migrate tool
docker exec apicore-api /app/apicore -migrate

# Method 2: Using cmd/migrate
docker run --rm \
  --network apicore_monitoring \
  -e DB_HOST=postgres \
  apicore:latest \
  go run cmd/migrate/main.go up
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

### Connect to Containers

```bash
# PostgreSQL
docker exec -it apicore-postgres psql -U postgres -d apicore

# Redis
docker exec -it apicore-redis redis-cli

# API container shell
docker exec -it apicore-api sh
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
- Input validation (cần thêm)
- Rate limiting (TODO)
- Authentication (TODO)
- Authorization (TODO)

## 🚀 Deployment

### Build binary

```bash
go build -o apicore cmd/app/main.go
```

### Run binary

```bash
./apicore
```

### Docker (TODO)

```dockerfile
FROM golang:1.23-alpine
WORKDIR /app
COPY . .
RUN go build -o apicore cmd/app/main.go
CMD ["./apicore"]
```

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

### Database Management

```bash
# Using Makefile
make migrate          # Run all pending migrations
make migrate-down     # Rollback all migrations
make migrate-version  # Show current version
make seed             # Run seeders

# Using migrate CLI (more options)
go run cmd/migrate/main.go up              # Run migrations
go run cmd/migrate/main.go down            # Rollback all
go run cmd/migrate/main.go version         # Show version
go run cmd/migrate/main.go steps -n 1      # Run 1 step
go run cmd/migrate/main.go force -version 1  # Force version
go run cmd/migrate/main.go seed            # Run seeders
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
docker-compose up -d postgres redis    # Infrastructure only
go run cmd/app/main.go                 # Run app locally

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

### Database Operations

```bash
# Connect to PostgreSQL
docker exec -it apicore-postgres psql -U postgres -d apicore

# Inside psql
\dt                   # List tables
\d users              # Describe users table
SELECT * FROM users;  # Query users
SELECT * FROM schema_migrations;  # Check migration version
```

### Logs

```bash
# Application logs
tail -f storages/logs/app.log

# Docker logs
make docker-logs                    # All services
docker logs -f apicore-api          # API only
docker logs -f apicore-postgres     # PostgreSQL
docker logs -f apicore-redis        # Redis
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

## 📝 TODO

- [x] ~~Thêm database support (PostgreSQL/MySQL)~~
- [x] ~~Thêm Docker support~~
- [x] ~~Thêm logging với structured logger (zerolog)~~
- [ ] Thêm authentication & authorization (JWT)
- [ ] Thêm validation với go-playground/validator
- [ ] Thêm unit tests
- [ ] Thêm integration tests
- [ ] Thêm rate limiting với Redis
- [ ] Thêm CORS support
- [ ] Thêm CI/CD pipeline (GitHub Actions)
- [ ] Thêm API versioning
- [ ] Thêm pagination
- [ ] Thêm filtering & sorting

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

## 👥 Authors

- Your Name - Initial work

## 🙏 Acknowledgments

- Chi router team
- Google Wire team
- Swagger/OpenAPI community
