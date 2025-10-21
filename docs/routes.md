# Cấu Trúc Routes

## Tổng Quan

Dự án sử dụng kiến trúc module-based routing với Chi router. Mỗi module có file `route.go` riêng để quản lý routes của nó.

## Cấu Trúc

```
internal/
├── routes/
│   └── routes.go          # File routes tổng - đăng ký tất cả modules
└── app/
    └── user/
        ├── route.go       # Routes riêng của module user
        ├── controller.go  # Handlers
        └── service.go     # Business logic
```

## Cách Thêm Module Mới

### Bước 1: Tạo cấu trúc module

Ví dụ thêm module `order`:

```
internal/app/order/
├── route.go
├── controller.go
└── service.go
```

### Bước 2: Tạo file route.go trong module

```go
package order

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
    r.Route("/orders", func(r chi.Router) {
        r.Get("/", h.Index)      // GET /api/v1/orders
        r.Post("/", h.Store)     // POST /api/v1/orders
        r.Get("/{id}", h.Show)   // GET /api/v1/orders/{id}
        r.Put("/{id}", h.Update) // PUT /api/v1/orders/{id}
        r.Delete("/{id}", h.Destroy) // DELETE /api/v1/orders/{id}
    })
}
```

### Bước 3: Cập nhật routes tổng

Trong `internal/routes/routes.go`:

```go
package routes

import (
    "anhnq/api-core/internal/app/order"
    "anhnq/api-core/internal/app/user"
    "github.com/go-chi/chi/v5"
)

type Controllers struct {
    UserHandler  *user.Handler
    OrderHandler *order.Handler  // Thêm handler mới
}

func RegisterRoutes(r chi.Router, c *Controllers) {
    r.Route("/api/v1", func(r chi.Router) {
        user.RegisterRoutes(r, c.UserHandler)
        order.RegisterRoutes(r, c.OrderHandler)  // Đăng ký routes mới
    })
}
```

### Bước 4: Cập nhật main.go

```go
// Khởi tạo repositories
orderRepo := repositories.NewOrderRepository()

// Khởi tạo services
orderService := order.NewService(orderRepo)

// Khởi tạo handlers
orderHandler := order.NewHandler(orderService)

// Đăng ký controllers
controllers := &routes.Controllers{
    UserHandler:  userHandler,
    OrderHandler: orderHandler,  // Thêm handler mới
}
```

## API Endpoints

### User Module

| Method | Endpoint             | Mô tả               |
| ------ | -------------------- | ------------------- |
| GET    | `/api/v1/users`      | Lấy danh sách users |
| POST   | `/api/v1/users`      | Tạo user mới        |
| GET    | `/api/v1/users/{id}` | Lấy user theo ID    |
| PUT    | `/api/v1/users/{id}` | Cập nhật user       |
| DELETE | `/api/v1/users/{id}` | Xóa user            |

### Health Check

| Method | Endpoint | Mô tả           |
| ------ | -------- | --------------- |
| GET    | `/ping`  | Kiểm tra server |

## Quy Ước

1. **Prefix**: Tất cả API routes đều có prefix `/api/v1`
2. **Module Routes**: Mỗi module có prefix riêng (ví dụ: `/users`, `/orders`)
3. **Handler Naming**:
   - `Index` - GET danh sách
   - `Show` - GET chi tiết
   - `Store` - POST tạo mới
   - `Update` - PUT cập nhật
   - `Destroy` - DELETE xóa

## Ví Dụ Request

### Tạo user mới

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

### Lấy danh sách users

```bash
curl http://localhost:8080/api/v1/users
```

### Lấy user theo ID

```bash
curl http://localhost:8080/api/v1/users/{id}
```

### Cập nhật user

```bash
curl -X PUT http://localhost:8080/api/v1/users/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com"
  }'
```

### Xóa user

```bash
curl -X DELETE http://localhost:8080/api/v1/users/{id}
```
