# Action Events Package

Package `pkg/actionEvent` cung cấp chức năng ghi action events vào Loki với job động và async push.

## Tính năng

- ✅ **Job động**: Truyền job name khác nhau cho mỗi lần gọi
- ✅ **Async push**: Không chờ response, chỉ push đi
- ✅ **Tích hợp Repository**: Tự động log CRUD operations
- ✅ **Structured JSON**: Dễ query và analyze
- ✅ **Performance tốt**: Timeout ngắn (2s) cho async operations
- ✅ **Silent fail**: Không block main operations nếu Loki down

## Cấu hình

Thêm vào `.env`:

```env
# Action Events
ACTION_EVENT_LOKI_URL=http://localhost:3100
ACTION_EVENT_ENVIRONMENT=development
ACTION_EVENT_ENABLED=true
ACTION_EVENT_DEFAULT_JOB=action_events
```

## Sử dụng

### 1. Tự động trong Repository

BaseRepository đã tích hợp sẵn với job động:

```go
// Tạo user - tự động log với job "user_create"
user := &models.User{Name: "John", Email: "john@example.com"}
err := userRepo.Create(ctx, user)

// Cập nhật user - tự động log với job "user_update"
user.Name = "John Updated"
err := userRepo.Update(ctx, user.ID, user)

// Xóa user - tự động log với job "user_delete"
err := userRepo.Delete(ctx, user.ID)
```

### 2. Manual Event Logging với Job Động

```go
// Login event với job "auth_events"
err := actionEvent.LogLogin(ctx, "auth_events", "user123", "192.168.1.1", "Mozilla/5.0...", map[string]interface{}{
    "login_method": "password",
    "success": true,
})

// Logout event với job "auth_events"
err := actionEvent.LogLogout(ctx, "auth_events", "user123", "192.168.1.1", "Mozilla/5.0...", map[string]interface{}{
    "session_duration": "2h30m",
})

// Create event với job "product_events"
err := actionEvent.LogCreate(ctx, "product_events", "product", "prod123", "user123", map[string]interface{}{
    "name": "New Product",
    "price": 99.99,
})

// Update event với job "order_events"
err := actionEvent.LogUpdate(ctx, "order_events", "order", "order123", "user123", map[string]interface{}{
    "status": "shipped",
})
```

### 3. Custom Event với Job Động

```go
event := actionEvent.Event{
    Action: "custom_action",
    Entity: "product",
    EntityID: "prod123",
    UserID: "user123",
    Data: map[string]interface{}{
        "custom_field": "value",
    },
    Job: "custom_events", // Dynamic job
}
err := actionEvent.LogEvent(ctx, event)
```

## Event Structure

```json
{
  "action": "create|update|delete|login|logout|custom",
  "entity": "user|product|order|...",
  "entity_id": "uuid-string",
  "user_id": "user-who-performed-action",
  "data": {
    "custom_fields": "values"
  },
  "timestamp": "2025-10-29T09:45:55Z",
  "ip": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "job": "dynamic_job_name"
}
```

## Job Naming Strategy

### Repository Auto Jobs

- `user_create`, `user_update`, `user_delete`
- `product_create`, `product_update`, `product_delete`
- `order_create`, `order_update`, `order_delete`

### Manual Jobs

- `auth_events` - Login/logout events
- `product_events` - Product-related events
- `order_events` - Order-related events
- `payment_events` - Payment-related events
- `notification_events` - Notification events

## Loki Query Examples

### Tìm tất cả create events

```logql
{job=~".*_create"} |= "create"
```

### Tìm events của job cụ thể

```logql
{job="auth_events"} | json
```

### Tìm events của user cụ thể

```logql
{job=~".*"} | json | user_id="user123"
```

### Tìm events của entity cụ thể

```logql
{job=~".*"} | json | entity="user"
```

### Tìm login events

```logql
{job="auth_events"} | json | action="login"
```

### Tìm events trong khoảng thời gian

```logql
{job=~".*"} | json | timestamp >= "2025-10-29T00:00:00Z"
```

## Performance Features

### Async Push

- Events được push async trong goroutine riêng
- Không chờ response từ Loki
- Timeout ngắn (2s) để không block quá lâu
- Silent fail nếu Loki không available

### Repository Integration

- Tự động detect entity name từ type
- Tự động tạo job name: `{entity}_{action}`
- Extract user_id từ context
- Chỉ log khi operation thành công

## Context Requirements

Để extract `user_id` từ context:

```go
ctx := context.WithValue(context.Background(), "user_id", "user123")
```

Hoặc trong middleware:

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract user from JWT
        userID := "user123"

        // Add to context
        ctx := context.WithValue(r.Context(), "user_id", userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Troubleshooting

1. **Events không được ghi**: Kiểm tra `ACTION_EVENT_ENABLED=true` và Loki server running
2. **Missing user_id**: Đảm bảo context có `user_id` key
3. **Connection timeout**: Kiểm tra `ACTION_EVENT_LOKI_URL` và network connectivity
4. **Job không hiển thị**: Kiểm tra job name được truyền đúng

## So sánh với Logger Package

| Feature     | Logger Package    | ActionEvent Package       |
| ----------- | ----------------- | ------------------------- |
| Purpose     | Debug/Logging     | Audit/Analytics           |
| Output      | Console/File/Loki | Chỉ Loki                  |
| Job         | Fixed             | Dynamic                   |
| Sync        | Sync              | Async                     |
| Performance | Có thể chậm       | Nhanh                     |
| Use Case    | Debug, Error logs | User actions, CRUD events |
