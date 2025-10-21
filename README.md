# 🚀 ApiCore - User Management API

API quản lý người dùng được xây dựng với Go, sử dụng kiến trúc modular và dependency injection.

## ✨ Tính Năng

- ✅ RESTful API design
- ✅ Kiến trúc modular với Chi Router
- ✅ Dependency Injection với Google Wire
- ✅ In-memory repository (dễ dàng mở rộng sang database)
- ✅ Swagger/OpenAPI documentation
- ✅ Interactive API documentation với Swagger UI
- ✅ Health check endpoint
- ✅ Request logging middleware
- ✅ Panic recovery middleware
- ✅ Request ID tracking

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
├── docs/
│   ├── index.html                  # Documentation home
│   ├── swagger.html                # Swagger UI
│   ├── swagger.json                # OpenAPI specification
│   ├── routes.md                   # Routes guide
│   └── swagger-guide.md            # Swagger usage guide
├── go.mod
└── go.sum
```

## 🚀 Quick Start

### Yêu Cầu

- Go 1.23.4 hoặc cao hơn
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

4. **Generate Wire code** (nếu cần)

```bash
wire ./internal/wire
```

5. **Chạy server**

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

| URL                      | Mô Tả                                |
| ------------------------ | ------------------------------------ |
| `/docs`                  | Trang chủ documentation với overview |
| `/swagger`               | Swagger UI - Interactive API testing |
| `/swagger.json`          | OpenAPI specification file           |
| `/docs/routes.md`        | Hướng dẫn về routes                  |
| `/docs/swagger-guide.md` | Hướng dẫn sử dụng Swagger            |

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

## 📦 Dependencies

- [chi](https://github.com/go-chi/chi) - Lightweight router
- [wire](https://github.com/google/wire) - Dependency injection
- [uuid](https://github.com/google/uuid) - UUID generation
- [zerolog](https://github.com/rs/zerolog) - Structured logging
- [loki-client-go](https://github.com/grafana/loki-client-go) - Loki integration

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

## 📝 TODO

- [ ] Thêm database support (PostgreSQL/MySQL)
- [ ] Thêm authentication & authorization
- [ ] Thêm validation với go-playground/validator
- [ ] Thêm unit tests
- [ ] Thêm integration tests
- [ ] Thêm rate limiting
- [ ] Thêm CORS support
- [ ] Thêm Docker support
- [ ] Thêm CI/CD pipeline
- [ ] Thêm logging với structured logger (zerolog/zap)

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
