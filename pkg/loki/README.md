# Loki Events Package

Package `pkg/loki` cung cấp chức năng ghi events vào Loki để audit và analytics.

## Tính năng

- ✅ Ghi events vào Loki với job label riêng biệt
- ✅ Chỉ ghi vào Loki (không ghi console/file)
- ✅ Structured JSON data cho query dễ dàng
- ✅ Async logging để không ảnh hưởng performance
- ✅ Tích hợp sẵn vào BaseRepository
- ✅ Helper functions cho các events phổ biến

## Cấu hình

Thêm vào `.env`:

```env
# Loki Events
LOKI_URL=http://localhost:3100
LOKI_JOB=action_events
LOKI_ENVIRONMENT=development
LOKI_ENABLED=true
```

## Sử dụng

### 1. Tự động trong Repository

BaseRepository đã tích hợp sẵn event logging:

```go
// Tạo user - tự động log create event
user := &models.User{Name: "John", Email: "john@example.com"}
err := userRepo.Create(ctx, user)

// Cập nhật user - tự động log update event
user.Name = "John Updated"
err := userRepo.Update(ctx, user.ID, user)

// Xóa user - tự động log delete event
err := userRepo.Delete(ctx, user.ID)
```

### 2. Manual Event Logging

```go
// Login event
err := loki.LogLogin(ctx, "user123", "192.168.1.1", "Mozilla/5.0...", map[string]interface{}{
    "login_method": "password",
    "success": true,
})

// Logout event
err := loki.LogLogout(ctx, "user123", "192.168.1.1", "Mozilla/5.0...", map[string]interface{}{
    "session_duration": "2h30m",
})

// Custom event
event := loki.Event{
    Action: "custom_action",
    Entity: "product",
    EntityID: "prod123",
    UserID: "user123",
    Data: map[string]interface{}{
        "custom_field": "value",
    },
}
err := loki.LogEvent(ctx, event)
```

### 3. Async Logging

```go
// Log async để không block operation
loki.LogEventAsync(ctx, event)
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
  "user_agent": "Mozilla/5.0..."
}
```

## Loki Query Examples

### Tìm tất cả create events

```logql
{job="action_events"} |= "create"
```

### Tìm events của user cụ thể

```logql
{job="action_events"} | json | user_id="user123"
```

### Tìm events của entity cụ thể

```logql
{job="action_events"} | json | entity="user"
```

### Tìm events trong khoảng thời gian

```logql
{job="action_events"} | json | timestamp >= "2025-10-29T00:00:00Z"
```

### Tìm login events

```logql
{job="action_events"} | json | action="login"
```

## Context Requirements

Để extract `user_id` từ context, đảm bảo context có key `user_id`:

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

## Performance

- Events được ghi async để không ảnh hưởng performance
- Timeout 5s cho HTTP requests
- Silent fail nếu Loki không available
- Không block main operations nếu Loki down

## Troubleshooting

1. **Events không được ghi**: Kiểm tra `LOKI_ENABLED=true` và Loki server running
2. **Missing user_id**: Đảm bảo context có `user_id` key
3. **Connection timeout**: Kiểm tra `LOKI_URL` và network connectivity
