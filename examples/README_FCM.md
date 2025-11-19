# Hướng Dẫn Test FCM (Firebase Cloud Messaging)

File này hướng dẫn cách test các tính năng FCM trong ApiCore.

## Chuẩn Bị

### 1. Tải Firebase Service Account File

1. Truy cập [Firebase Console](https://console.firebase.google.com/)
2. Chọn project của bạn
3. Vào **Settings** (⚙️) > **Project settings**
4. Chọn tab **Service accounts**
5. Click **Generate new private key**
6. Lưu file JSON vào thư mục `keys/` với tên `firebase-credentials.json`

   ```
   keys/
   └── firebase-credentials.json
   ```

### 2. Cấu Trúc File Service Account

File `firebase-credentials.json` có cấu trúc như sau:

```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "xxx",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "firebase-adminsdk-xxx@your-project-id.iam.gserviceaccount.com",
  "client_id": "xxx",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/..."
}
```

**⚠️ Lưu ý quan trọng:**

- File này chứa thông tin bảo mật (private key), **KHÔNG** commit vào Git
- Đã được thêm vào `.gitignore` để tránh commit nhầm
- Chỉ sử dụng trong môi trường development/testing
- Trong production, nên sử dụng environment variables hoặc secret management

### 3. Lấy Firebase Web Config (Cho HTML Test)

Để test bằng HTML interface, cần có Firebase Web Config:

1. Truy cập [Firebase Console](https://console.firebase.google.com/)
2. Chọn project của bạn
3. Vào **Settings** (⚙️) > **General**
4. Scroll xuống phần **Your apps**
5. Nếu chưa có Web app:
   - Click **Add app** > Chọn **Web** (</>)
   - Đặt tên app (ví dụ: "FCM Test")
   - Click **Register app**
6. Copy config object (có dạng JSON)
7. Paste vào file `test_fcm.html`

### 4. Lấy Device Token

Có nhiều cách để lấy device token:

**Cách 1: Sử dụng HTML Interface (Dễ nhất)**

- Mở `examples/test_fcm.html`
- Khởi tạo Firebase và yêu cầu quyền
- Token sẽ tự động hiển thị và có thể copy

**Cách 2: Từ ứng dụng mobile/web:**

- **Android**: Sử dụng Firebase SDK để lấy token
- **iOS**: Sử dụng Firebase SDK để lấy token
- **Web**: Sử dụng Firebase SDK để lấy token

## Cách Sử Dụng

### Option 1: Test Bằng HTML Interface (Khuyến Nghị)

Cách đơn giản nhất để test FCM với giao diện trực quan:

1. **Mở file HTML:**

   ```bash
   # Mở file trong browser
   open examples/test_fcm.html
   # hoặc
   firefox examples/test_fcm.html
   # hoặc
   chrome examples/test_fcm.html
   ```

2. **Lấy Firebase Web Config:**

   - Truy cập [Firebase Console](https://console.firebase.google.com/)
   - Chọn project của bạn
   - Vào **Settings** > **General**
   - Scroll xuống phần **Your apps**
   - Nếu chưa có Web app, click **Add app** > **Web** (</>) để tạo
   - Copy toàn bộ config object (JSON)

3. **Paste config vào file HTML** và click "Khởi Tạo Firebase"

4. **Yêu cầu quyền notification** từ browser

5. **Lấy FCM token** - Token sẽ hiển thị và có thể copy

6. **Test nhận notification:**
   - Sử dụng token để gửi notification từ backend
   - Hoặc subscribe vào topic và gửi notification đến topic

**Tính năng của HTML Interface:**

- ✅ Khởi tạo Firebase dễ dàng
- ✅ Lấy và copy FCM token
- ✅ Hiển thị notifications nhận được
- ✅ Test gửi notification từ backend
- ✅ Subscribe/Unsubscribe topic
- ✅ Giao diện đẹp, dễ sử dụng

**Lưu ý về Service Worker:**

- Để nhận notification khi tab đang đóng, cần cấu hình Service Worker
- Copy file `firebase-messaging-sw.js` vào thư mục public của website
- Cập nhật Firebase config trong file Service Worker
- Service Worker cần được serve qua HTTPS (hoặc localhost)

### Option 2: Chạy Test Script (Go)

```bash
cd examples

# Xem tất cả các options
go run test_fcm.go -h

# Test tất cả tính năng
go run test_fcm.go -test=all

# Gửi notification đến một device token
go run test_fcm.go -test=token -token=YOUR_DEVICE_TOKEN

# Gửi notification đến nhiều tokens
go run test_fcm.go -test=tokens -tokens=TOKEN1,TOKEN2,TOKEN3

# Gửi notification đến topic
go run test_fcm.go -test=topic -topic=news

# Gửi notification với condition
go run test_fcm.go -test=condition -condition="'news' in topics || 'sports' in topics"

# Subscribe tokens vào topic
go run test_fcm.go -test=subscribe -tokens=TOKEN1,TOKEN2 -topic=news

# Unsubscribe tokens khỏi topic
go run test_fcm.go -test=unsubscribe -tokens=TOKEN1,TOKEN2 -topic=news

# Dry run (test mà không gửi thực sự)
go run test_fcm.go -test=dryrun -token=YOUR_DEVICE_TOKEN

# Sử dụng credentials file khác
go run test_fcm.go -credentials=keys/custom-credentials.json -test=token -token=YOUR_TOKEN
```

## Các Tính Năng Test

Script test bao gồm các tính năng sau:

### 1. **SendToToken** - Gửi đến một token

Gửi notification đến một device token cụ thể.

```bash
go run test_fcm.go -test=token -token=YOUR_DEVICE_TOKEN
```

### 2. **SendToTokens** - Gửi đến nhiều tokens (Multicast)

Gửi notification đến nhiều device tokens cùng lúc (tối đa 500 tokens).

```bash
go run test_fcm.go -test=tokens -tokens=TOKEN1,TOKEN2,TOKEN3
```

### 3. **SendToTopic** - Gửi đến topic

Gửi notification đến tất cả devices đã subscribe một topic.

```bash
go run test_fcm.go -test=topic -topic=news
```

### 4. **SendToCondition** - Gửi với condition

Gửi notification dựa trên điều kiện topic (AND, OR).

```bash
go run test_fcm.go -test=condition -condition="'news' in topics && 'sports' in topics"
```

### 5. **SubscribeToTopic** - Đăng ký topic

Đăng ký device tokens vào một topic.

```bash
go run test_fcm.go -test=subscribe -tokens=TOKEN1,TOKEN2 -topic=news
```

### 6. **UnsubscribeFromTopic** - Hủy đăng ký topic

Hủy đăng ký device tokens khỏi một topic.

```bash
go run test_fcm.go -test=unsubscribe -tokens=TOKEN1,TOKEN2 -topic=news
```

### 7. **SendDryRun** - Test không gửi thực sự

Test validation mà không gửi notification thực sự.

```bash
go run test_fcm.go -test=dryrun -token=YOUR_DEVICE_TOKEN
```

### 8. **Android Custom Config**

Gửi notification với cấu hình Android tùy chỉnh:

- Priority: high/normal
- Sound: custom sound
- Color: notification color
- Icon: custom icon
- TTL: time to live
- Click action

### 9. **iOS Custom Config**

Gửi notification với cấu hình iOS tùy chỉnh:

- Badge: badge count
- Sound: custom sound
- Category: notification category
- Thread ID: thread identifier
- Content Available: background sync
- Mutable Content: extension support

### 10. **Data-Only Notification** (Silent)

Gửi data mà không hiển thị notification trên thiết bị.

### 11. **SendAll** - Gửi nhiều messages khác nhau

Gửi nhiều messages khác nhau đến nhiều tokens cùng lúc (tối đa 500 messages).

## Examples

### Example 1: Gửi notification đơn giản

```bash
go run test_fcm.go \
  -test=token \
  -token=YOUR_DEVICE_TOKEN
```

### Example 2: Gửi notification với Android custom config

Script tự động test tính năng này khi chạy `-test=all`.

### Example 3: Gửi notification đến topic

```bash
go run test_fcm.go \
  -test=topic \
  -topic=breaking_news
```

### Example 4: Subscribe và gửi notification

```bash
# Bước 1: Subscribe tokens vào topic
go run test_fcm.go \
  -test=subscribe \
  -tokens=TOKEN1,TOKEN2 \
  -topic=news

# Bước 2: Gửi notification đến topic
go run test_fcm.go \
  -test=topic \
  -topic=news
```

### Example 5: Gửi notification với condition

```bash
# Gửi đến users đăng ký cả "news" VÀ "sports"
go run test_fcm.go \
  -test=condition \
  -condition="'news' in topics && 'sports' in topics"

# Gửi đến users đăng ký "news" HOẶC "sports"
go run test_fcm.go \
  -test=condition \
  -condition="'news' in topics || 'sports' in topics"
```

## Xử Lý Lỗi

### Lỗi thường gặp:

1. **File credentials không tồn tại**

   ```
   ❌ Không tìm thấy file credentials: keys/firebase-credentials.json
   ```

   → Kiểm tra file đã được đặt đúng vị trí chưa

2. **Invalid registration token**

   ```
   ❌ Lỗi: invalid-registration-token
   ```

   → Token không hợp lệ hoặc đã hết hạn. Cần lấy token mới từ app

3. **Token không đăng ký**

   ```
   ❌ Lỗi: registration-token-not-registered
   ```

   → Token không còn tồn tại trong Firebase. Cần xóa khỏi database

4. **Permission denied**
   ```
   ❌ Lỗi: permission-denied
   ```
   → Service account không có quyền. Kiểm tra quyền trong Firebase Console

## Best Practices

1. **Sử dụng Dry Run để test trước**

   ```bash
   go run test_fcm.go -test=dryrun -token=YOUR_TOKEN
   ```

2. **Validate tokens trước khi gửi**

   - Sử dụng dry-run để kiểm tra token hợp lệ
   - Xóa invalid tokens khỏi database

3. **Xử lý lỗi từ Batch Response**

   - Kiểm tra `response.SuccessCount` và `response.FailureCount`
   - Xử lý từng lỗi trong `response.Responses`

4. **Sử dụng Topics cho broadcast**

   - Thay vì gửi đến nhiều tokens, sử dụng topics
   - Hiệu quả hơn và dễ quản lý

5. **Bảo mật Service Account File**
   - Không commit vào Git
   - Sử dụng environment variables trong production
   - Rotate keys định kỳ

## File HTML Test Interface

### Cấu Trúc File

- `test_fcm.html` - Giao diện HTML để test FCM
- `firebase-messaging-sw.js` - Service Worker để nhận notification khi tab đóng

### Sử Dụng HTML Interface

1. **Mở file `test_fcm.html` trong browser**

2. **Nhập Firebase Web Config:**

   ```json
   {
     "apiKey": "AIza...",
     "authDomain": "your-project.firebaseapp.com",
     "projectId": "your-project-id",
     "storageBucket": "your-project.appspot.com",
     "messagingSenderId": "123456789",
     "appId": "1:123456789:web:abc123"
   }
   ```

3. **Click "Khởi Tạo Firebase"**

4. **Yêu cầu quyền notification** - Browser sẽ hỏi quyền

5. **Lấy token** - Click "Lấy FCM Token", token sẽ hiển thị

6. **Copy token** để sử dụng cho backend hoặc test

7. **Test notification:**
   - Nhập token vào Go test script
   - Hoặc gửi từ backend API (cần tạo endpoint)

### Cấu Hình Service Worker (Tùy Chọn)

Để nhận notification khi tab đang đóng:

1. Copy `firebase-messaging-sw.js` vào thư mục public của website
2. Cập nhật Firebase config trong file Service Worker
3. Service Worker sẽ tự động được register

**Lưu ý:**

- Service Worker chỉ hoạt động qua HTTPS hoặc localhost
- Cần cấu hình VAPID key nếu muốn custom notification

## Xem Thêm

- [FCM Package README](../../pkg/fcm/README.md)
- [Firebase Cloud Messaging Documentation](https://firebase.google.com/docs/cloud-messaging)
- [Firebase Admin Go SDK](https://firebase.google.com/docs/admin/setup)
- [Firebase Web Setup](https://firebase.google.com/docs/web/setup)
