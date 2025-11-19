# Ví Dụ Sử Dụng FCM trong User Service

File này hướng dẫn cách sử dụng FCM client trong user service để gửi push notification.

## Cấu Hình

FCM client đã được inject vào User Service thông qua Wire dependency injection. Client sẽ là `nil` nếu không được cấu hình (file credentials không tồn tại hoặc lỗi khởi tạo).

## Các Method Có Sẵn

### 1. SendNotificationToToken

Gửi notification đến một FCM token cụ thể:

```go
// Trong controller hoặc service method
func (s *Service) SomeMethod(ctx context.Context) error {
    // Lấy token từ request, database, hoặc từ user
    token := "USER_FCM_TOKEN_HERE"

    // Gửi notification
    messageID, err := s.SendNotificationToToken(
        ctx,
        token,
        "Tiêu đề thông báo",
        "Nội dung thông báo",
        map[string]string{
            "type":      "order_update",
            "order_id":  "12345",
            "action":    "view_order",
            "deep_link": "app://orders/12345",
        },
    )

    if err != nil {
        // Xử lý lỗi
        return err
    }

    // Log success
    fmt.Printf("Notification sent: %s\n", messageID)
    return nil
}
```

### 2. SendNotificationToTokens

Gửi notification đến nhiều tokens cùng lúc (multicast):

```go
// Lấy danh sách tokens từ database
tokens := []string{
    "token1",
    "token2",
    "token3",
}

successCount, failureCount, err := s.SendNotificationToTokens(
    ctx,
    tokens,
    "Thông báo chung",
    "Gửi đến tất cả users",
    map[string]string{
        "type": "announcement",
    },
)

if err != nil {
    return err
}

fmt.Printf("Success: %d, Failed: %d\n", successCount, failureCount)
```

### 3. SendNotificationToUser

Gửi notification đến user (cần implement logic lấy token từ DB):

```go
// TODO: Implement getUserFCMToken() trong service
userID := uuid.MustParse("user-uuid-here")

err := s.SendNotificationToUser(
    ctx,
    userID,
    "Xin chào!",
    "Bạn có thông báo mới",
    map[string]string{
        "type": "greeting",
    },
)

if err != nil {
    return err
}
```

## Ví Dụ Tích Hợp Vào Business Logic

### Ví dụ 1: ✅ ĐÃ IMPLEMENT - Gửi notification khi tạo user mới

**Đã được tích hợp sẵn trong `Service.Create()` method:**

```go
// Khi tạo user mới, notification sẽ tự động được gửi nếu có FCM token
// POST /api/v1/users
{
    "name": "Nguyễn Văn A",
    "email": "nguyenvana@example.com",
    "fcm_token": "USER_FCM_TOKEN_HERE" // Optional
}
```

**Flow hoạt động:**

1. User được tạo thành công
2. Nếu có `fcm_token` trong request → Gửi notification chào mừng tự động
3. Notification được gửi trong background (không block response)
4. Nếu không có token → Bỏ qua (không lỗi)

**Notification sẽ chứa:**

- Title: "Chào mừng đến với ApiCore!"
- Body: "Xin chào {name}! Tài khoản của bạn đã được tạo thành công."
- Data:
  ```json
  {
    "type": "user_created",
    "user_id": "uuid",
    "email": "email",
    "action": "view_profile",
    "deep_link": "app://users/{id}",
    "timestamp": "2025-01-30T..."
  }
  ```

**Code thực tế trong service:**

```go
// internal/app/user/service.go - Create method
func (s *Service) Create(ctx context.Context, user model.User, avatarFile *multipart.FileHeader, fcmToken ...string) (*model.User, error) {
    // ... tạo user ...

    // Gửi FCM notification chào mừng (background)
    var token string
    if len(fcmToken) > 0 && fcmToken[0] != "" {
        token = fcmToken[0]
    }
    go s.sendWelcomeNotification(context.Background(), &user, token)

    return &user, nil
}
```

### Ví dụ 2: Gửi notification thủ công trong service method

```go
// Ví dụ: Gửi notification khi user cập nhật profile
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, data UpdateData) error {
    // ... update logic ...

    // Lấy FCM token của user từ database
    user, err := s.repo.FindByID(ctx, userID)
    if err != nil {
        return err
    }

    // Gửi notification (nếu có token)
    if user.FCMToken != nil && *user.FCMToken != "" {
        go func() {
            s.SendNotificationToToken(
                context.Background(),
                *user.FCMToken,
                "Profile đã được cập nhật",
                "Thông tin của bạn đã được cập nhật thành công",
                map[string]string{
                    "type":    "profile_updated",
                    "user_id": userID.String(),
                },
            )
        }()
    }

    return nil
}
```

### Ví dụ 3: Gửi notification khi có thông báo đơn hàng

```go
// Trong order service (nếu có) hoặc notification service
func (s *Service) NotifyOrderUpdate(ctx context.Context, userID uuid.UUID, orderID string, status string) error {
    // Lấy FCM token của user
    user, err := s.repo.FindByID(ctx, userID)
    if err != nil {
        return err
    }

    // Kiểm tra user có token không
    // if user.FCMToken == nil || *user.FCMToken == "" {
    //     return fmt.Errorf("user does not have FCM token")
    // }

    // Gửi notification
    _, err = s.SendNotificationToToken(
        ctx,
        *user.FCMToken, // hoặc lấy từ device_tokens table
        "Cập nhật đơn hàng",
        fmt.Sprintf("Đơn hàng #%s đã được %s", orderID, status),
        map[string]string{
            "type":     "order_update",
            "order_id": orderID,
            "status":   status,
            "action":   "view_order",
            "deep_link": fmt.Sprintf("app://orders/%s", orderID),
        },
    )

    return err
}
```

### Ví dụ 3: Gửi notification đến nhiều users (broadcast)

```go
func (s *Service) BroadcastAnnouncement(ctx context.Context, title, body string) error {
    // Lấy tất cả FCM tokens từ database
    // tokens, err := s.getAllFCMTokens(ctx)

    // Hoặc lấy từ một nhóm users cụ thể
    users, err := s.repo.FindAll(ctx)
    if err != nil {
        return err
    }

    var tokens []string
    for _, user := range users {
        // if user.FCMToken != nil && *user.FCMToken != "" {
        //     tokens = append(tokens, *user.FCMToken)
        // }
    }

    if len(tokens) == 0 {
        return fmt.Errorf("no FCM tokens available")
    }

    // Chia thành batch 500 tokens (giới hạn của FCM)
    batchSize := 500
    for i := 0; i < len(tokens); i += batchSize {
        end := i + batchSize
        if end > len(tokens) {
            end = len(tokens)
        }

        batch := tokens[i:end]
        successCount, failureCount, err := s.SendNotificationToTokens(
            ctx,
            batch,
            title,
            body,
            map[string]string{
                "type": "announcement",
            },
        )

        if err != nil {
            // Log error nhưng tiếp tục với batch tiếp theo
            fmt.Printf("Batch %d failed: %v\n", i/batchSize, err)
            continue
        }

        fmt.Printf("Batch %d: Success=%d, Failed=%d\n", i/batchSize, successCount, failureCount)
    }

    return nil
}
```

## Lưu Ý Quan Trọng

1. **FCM Client có thể nil**: Luôn kiểm tra `if s.fcmClient == nil` hoặc sử dụng các method có sẵn đã có check

2. **Lưu FCM Token**: Cần có cơ chế lưu FCM token vào database:

   - Thêm field `fcm_token` vào bảng `users`
   - Hoặc tạo bảng `device_tokens` riêng để lưu nhiều tokens per user

3. **Xử lý lỗi Invalid Token**: Khi gửi thất bại do invalid token, nên xóa token khỏi database:

   ```go
   if strings.Contains(err.Error(), "invalid-registration-token") ||
      strings.Contains(err.Error(), "registration-token-not-registered") {
       // Xóa token khỏi database
       s.deleteFCMToken(ctx, token)
   }
   ```

4. **Background Processing**: Nên gửi notification trong goroutine để không block request:

   ```go
   go func() {
       // Send notification
   }()
   ```

5. **Error Handling**: Không nên fail business logic nếu notification fail:
   ```go
   if err := s.SendNotification(...); err != nil {
       // Log error nhưng không return
       logger.Errorf("Failed to send notification: %v", err)
   }
   ```

## Migration Database (Tùy Chọn)

Để lưu FCM token, có thể thêm vào bảng users:

```sql
ALTER TABLE users ADD COLUMN fcm_token TEXT;
CREATE INDEX idx_users_fcm_token ON users(fcm_token) WHERE fcm_token IS NOT NULL;
```

Hoặc tạo bảng device_tokens riêng:

```sql
CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    platform VARCHAR(10) NOT NULL, -- 'ios', 'android', 'web'
    device_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP
);

CREATE INDEX idx_device_tokens_user_id ON device_tokens(user_id);
CREATE INDEX idx_device_tokens_token ON device_tokens(token);
```
