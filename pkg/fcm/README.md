# FCM Package

Package FCM cung cấp các công cụ để gửi Firebase Cloud Messaging (Push Notifications) đến các thiết bị iOS, Android và Web.

## Cài đặt

Thêm Firebase Admin SDK vào project:

```bash
go get firebase.google.com/go/v4
go get google.golang.org/api/option
```

## Cấu hình

### 1. Lấy Firebase Credentials

1. Truy cập [Firebase Console](https://console.firebase.google.com/)
2. Chọn project của bạn
3. Vào Settings > Service accounts
4. Click "Generate new private key"
5. Lưu file JSON vào thư mục `keys/` trong project (ví dụ: `keys/firebase-credentials.json`)

### 2. Cập nhật Environment Variables

Thêm vào file `.env`:

```env
# Firebase Configuration
FIREBASE_CREDENTIALS_FILE=keys/firebase-credentials.json
FCM_TIMEOUT=10
```

## Sử dụng

### Khởi tạo FCM Client

```go
package main

import (
    "anhnq/api-core/pkg/fcm"
    "time"
)

func main() {
    // Tạo config
    config := &fcm.Config{
        CredentialsFile: "keys/firebase-credentials.json",
        Timeout:         10 * time.Second,
    }

    // Khởi tạo client
    client, err := fcm.NewClient(config)
    if err != nil {
        panic(err)
    }

    // Sử dụng client...
}
```

### Gửi Notification Đơn Giản

```go
// Tạo notification
notification := fcm.NewNotificationBuilder().
    SetTitle("Xin chào!").
    SetBody("Đây là thông báo đầu tiên của bạn").
    Build()

// Gửi đến device token
messageID, err := client.SendToToken(
    context.Background(),
    "device_token_here",
    notification,
    nil, // data (optional)
)
```

### Gửi Notification Kèm Data

```go
notification := fcm.NewNotificationBuilder().
    SetTitle("Đơn hàng mới").
    SetBody("Bạn có một đơn hàng mới #12345").
    Build()

data := map[string]string{
    "order_id": "12345",
    "user_id":  "67890",
    "action":   "view_order",
}

messageID, err := client.SendToToken(
    context.Background(),
    deviceToken,
    notification,
    data,
)
```

### Gửi Notification Với Hình Ảnh

```go
notification := fcm.NewNotificationBuilder().
    SetTitle("Sản phẩm mới").
    SetBody("Khám phá sản phẩm mới của chúng tôi!").
    SetImageURL("https://example.com/product-image.jpg").
    Build()

messageID, err := client.SendToToken(ctx, token, notification, nil)
```

### Tùy Chỉnh Cho Android

```go
notification := fcm.NewNotificationBuilder().
    SetTitle("Android Notification").
    SetBody("Notification với cấu hình Android tùy chỉnh").
    WithAndroidPriority("high").              // Priority cao
    WithAndroidSound("notification_sound").   // Custom sound
    WithAndroidColor("#FF5722").              // Màu notification
    WithAndroidIcon("ic_notification").       // Custom icon
    WithAndroidTTL(3600).                     // TTL 1 giờ (3600 giây)
    Build()
```

### Tùy Chỉnh Cho iOS

```go
notification := fcm.NewNotificationBuilder().
    SetTitle("iOS Notification").
    SetBody("Notification với cấu hình iOS tùy chỉnh").
    WithIOSBadge(5).                   // Badge count
    WithIOSSound("notification.wav").  // Custom sound
    WithIOSCategory("NEW_MESSAGE").    // Category
    WithIOSThreadID("thread-123").     // Thread ID
    Build()
```

### Gửi Đến Topic

```go
notification := fcm.NewNotificationBuilder().
    SetTitle("Breaking News").
    SetBody("Tin tức mới nhất từ chúng tôi").
    Build()

messageID, err := client.SendToTopic(
    context.Background(),
    "news", // topic name
    notification,
    nil,
)
```

### Gửi Đến Nhiều Tokens

```go
tokens := []string{
    "token1",
    "token2",
    "token3",
}

notification := fcm.NewNotificationBuilder().
    SetTitle("Bulk Notification").
    SetBody("Gửi đến nhiều devices").
    Build()

response, err := client.SendToTokens(ctx, tokens, notification, nil)

fmt.Printf("Success: %d, Failed: %d\n",
    response.SuccessCount,
    response.FailureCount,
)

// Kiểm tra các lỗi
for i, resp := range response.Responses {
    if !resp.Success {
        fmt.Printf("Token %s failed: %v\n", tokens[i], resp.Error)
    }
}
```

### Gửi Với Condition

```go
// Gửi đến users đăng ký cả "news" VÀ "sports"
condition := "'news' in topics && 'sports' in topics"

messageID, err := client.SendToCondition(ctx, condition, notification, nil)

// Gửi đến users đăng ký "news" HOẶC "sports"
condition = "'news' in topics || 'sports' in topics"
```

### Subscribe/Unsubscribe Topic

```go
// Subscribe tokens vào topic
tokens := []string{"token1", "token2"}

response, err := client.SubscribeToTopic(ctx, tokens, "news")
fmt.Printf("Success: %d\n", response.SuccessCount)

// Unsubscribe tokens khỏi topic
response, err := client.UnsubscribeFromTopic(ctx, tokens, "news")
```

### Silent Notification (Data-only)

```go
// Gửi data mà không hiển thị notification
data := map[string]string{
    "type":    "sync",
    "sync_id": "abc123",
}

messageID, err := client.SendToToken(ctx, token, nil, data)
```

### Dry Run (Test)

```go
// Test notification mà không gửi thật
notification := fcm.NewNotificationBuilder().
    SetTitle("Test").
    SetBody("This is a test").
    Build()

messageID, err := client.SendDryRun(ctx, token, notification, nil)
```

## Notification Builder

Package cung cấp `NotificationBuilder` để xây dựng notification dễ dàng:

```go
notification := fcm.NewNotificationBuilder().
    SetTitle("Title").
    SetBody("Body").
    SetImageURL("https://...").

    // Android specific
    WithAndroidPriority("high").
    WithAndroidSound("default").
    WithAndroidColor("#FF5722").
    WithAndroidIcon("ic_notification").
    WithAndroidClickAction("OPEN_ACTIVITY").
    WithAndroidTTL(3600).

    // iOS specific
    WithIOSBadge(10).
    WithIOSSound("default").
    WithIOSCategory("MESSAGE").
    WithIOSThreadID("thread-1").
    WithIOSContentAvailable().
    WithIOSMutableContent().

    Build()
```

## Best Practices

### 1. Error Handling

```go
messageID, err := client.SendToToken(ctx, token, notification, data)
if err != nil {
    // Log error
    logger.ErrorWithErr(err, "Failed to send FCM notification")

    // Kiểm tra loại lỗi
    if strings.Contains(err.Error(), "invalid-registration-token") {
        // Token không hợp lệ, xóa khỏi database
    } else if strings.Contains(err.Error(), "registration-token-not-registered") {
        // Token không còn tồn tại, xóa khỏi database
    }

    return err
}
```

### 2. Context Timeout

```go
// Tạo context với timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

messageID, err := client.SendToToken(ctx, token, notification, data)
```

### 3. Batch Operations

```go
// Chia tokens thành các batch 500 tokens
const batchSize = 500

for i := 0; i < len(allTokens); i += batchSize {
    end := i + batchSize
    if end > len(allTokens) {
        end = len(allTokens)
    }

    batch := allTokens[i:end]
    response, err := client.SendToTokens(ctx, batch, notification, data)
    if err != nil {
        logger.ErrorWithErr(err, "Failed to send batch")
        continue
    }

    // Xử lý response...
}
```

### 4. Lưu Device Token

Khi nhận device token từ client app, lưu vào database:

```sql
CREATE TABLE device_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token TEXT NOT NULL UNIQUE,
    platform VARCHAR(10) NOT NULL, -- 'ios', 'android', 'web'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_used_at TIMESTAMP
);

CREATE INDEX idx_device_tokens_user_id ON device_tokens(user_id);
CREATE INDEX idx_device_tokens_token ON device_tokens(token);
```

### 5. Xử Lý Invalid Tokens

```go
func sendNotificationAndCleanupInvalidTokens(
    fcmClient *fcm.Client,
    tokens []string,
    notification *fcm.Notification,
) {
    response, err := fcmClient.SendToTokens(ctx, tokens, notification, nil)
    if err != nil {
        return
    }

    // Tìm và xóa các invalid tokens
    var invalidTokens []string
    for i, resp := range response.Responses {
        if !resp.Success {
            errMsg := resp.Error.Error()
            if strings.Contains(errMsg, "invalid") ||
               strings.Contains(errMsg, "not-registered") {
                invalidTokens = append(invalidTokens, tokens[i])
            }
        }
    }

    // Xóa invalid tokens khỏi database
    if len(invalidTokens) > 0 {
        deleteInvalidTokens(invalidTokens)
    }
}
```

## Xem Thêm

- [Firebase Cloud Messaging Documentation](https://firebase.google.com/docs/cloud-messaging)
- [Firebase Admin Go SDK](https://firebase.google.com/docs/admin/setup)
- [FCM HTTP v1 API Reference](https://firebase.google.com/docs/reference/fcm/rest/v1/projects.messages)
