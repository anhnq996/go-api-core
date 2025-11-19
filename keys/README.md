# Firebase Service Account File

Thư mục này chứa file credentials của Firebase Service Account dùng cho FCM (Firebase Cloud Messaging).

## ⚠️ QUAN TRỌNG - Bảo Mật

**THÔNG TIN NHẠY CẢM - KHÔNG COMMIT VÀO GIT**

File trong thư mục này chứa:

- Private key của Firebase Service Account
- Thông tin xác thực để truy cập Firebase Admin API

File này **ĐÃ ĐƯỢC** thêm vào `.gitignore` để tránh commit nhầm.

## Cách Lấy File Service Account

### Bước 1: Truy cập Firebase Console

1. Đi đến [Firebase Console](https://console.firebase.google.com/)
2. Chọn project của bạn hoặc tạo project mới

### Bước 2: Tạo Service Account

1. Click vào biểu tượng ⚙️ (Settings) ở góc trên bên trái
2. Chọn **Project settings**
3. Chọn tab **Service accounts**
4. Trong phần "Admin SDK configuration snippet", chọn **Go** (hoặc bất kỳ ngôn ngữ nào)
5. Click nút **Generate new private key**
6. Một dialog sẽ xuất hiện, click **Generate key**
7. File JSON sẽ được tự động download về máy

### Bước 3: Lưu File Vào Project

1. Đổi tên file download thành `firebase-credentials.json`
2. Copy file vào thư mục `keys/` trong project:

   ```
   keys/
   └── firebase-credentials.json
   ```

### Bước 4: Xác Nhận File Đã Được Thêm Vào .gitignore

Kiểm tra file `.gitignore` có dòng:

```
keys/*.json
keys/firebase-credentials.json
```

Nếu chưa có, thêm vào để đảm bảo không commit nhầm.

## Cấu Trúc File Service Account

File `firebase-credentials.json` có cấu trúc như sau:

```json
{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "abc123def456...",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...\n-----END PRIVATE KEY-----\n",
  "client_email": "firebase-adminsdk-xxxxx@your-project-id.iam.gserviceaccount.com",
  "client_id": "123456789",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-xxxxx%40your-project-id.iam.gserviceaccount.com"
}
```

### Các Trường Quan Trọng:

- **`type`**: Luôn là `"service_account"`
- **`project_id`**: ID của Firebase project
- **`private_key`**: Private key dùng để xác thực (NHẠY CẢM - không chia sẻ)
- **`client_email`**: Email của service account
- **`client_id`**: ID của service account

## Quyền Cần Thiết

Service account cần có quyền:

- **Firebase Cloud Messaging API Admin**: Để gửi notifications
- **Firebase Admin**: Để quản lý FCM

Các quyền này thường được cấp tự động khi tạo service account từ Firebase Console.

## Sử Dụng Trong Code

### Cách 1: Sử dụng file trực tiếp

```go
import "api-core/pkg/fcm"

config := &fcm.Config{
    CredentialsFile: "keys/firebase-credentials.json",
    Timeout:         10 * time.Second,
}

client, err := fcm.NewClient(config)
```

### Cách 2: Sử dụng environment variable

```go
credentialsFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")
if credentialsFile == "" {
    credentialsFile = "keys/firebase-credentials.json" // default
}

config := &fcm.Config{
    CredentialsFile: credentialsFile,
    Timeout:         10 * time.Second,
}
```

Thêm vào file `.env`:

```env
FIREBASE_CREDENTIALS_FILE=keys/firebase-credentials.json
```

### Cách 3: Production - Sử dụng Base64 hoặc Secret Management

Trong production, không nên lưu file trực tiếp. Có thể:

1. **Lưu trong environment variable dạng Base64:**

   ```bash
   export FIREBASE_CREDENTIALS_BASE64=$(cat keys/firebase-credentials.json | base64)
   ```

2. **Sử dụng Secret Management Service:**
   - AWS Secrets Manager
   - Google Cloud Secret Manager
   - HashiCorp Vault
   - etc.

## Xác Thực Hoạt Động

Kiểm tra file credentials hoạt động đúng:

```bash
# Chạy test FCM
cd examples
go run test_fcm.go -test=dryrun -token=TEST_TOKEN
```

Nếu không có lỗi "permission denied" hoặc "invalid credentials", file đã hoạt động đúng.

## Troubleshooting

### Lỗi: "cannot find file"

```
❌ Không tìm thấy file credentials: keys/firebase-credentials.json
```

**Giải pháp:**

- Kiểm tra file đã được đặt đúng vị trí
- Kiểm tra quyền đọc file: `chmod 644 keys/firebase-credentials.json`
- Kiểm tra đường dẫn tuyệt đối nếu cần

### Lỗi: "permission denied"

```
❌ Lỗi: permission-denied
```

**Giải pháp:**

- Kiểm tra service account có quyền FCM Admin trong Firebase Console
- Kiểm tra project ID đúng
- Tạo lại service account mới nếu cần

### Lỗi: "invalid credentials"

```
❌ Lỗi: invalid credentials
```

**Giải pháp:**

- Kiểm tra file JSON hợp lệ: `cat keys/firebase-credentials.json | jq .`
- Kiểm tra private_key còn đầy đủ (có `\n` trong JSON)
- Tải lại file từ Firebase Console

### Lỗi: "project not found"

```
❌ Lỗi: project not found
```

**Giải pháp:**

- Kiểm tra `project_id` trong file JSON đúng
- Kiểm tra project còn tồn tại trong Firebase Console

## Rotate Keys (Xoay Keys)

Nên xoay keys định kỳ để bảo mật:

1. Tạo service account mới trong Firebase Console
2. Download file JSON mới
3. Thay thế file cũ: `keys/firebase-credentials.json`
4. Test lại để đảm bảo hoạt động
5. Xóa service account cũ trong Firebase Console (nếu không cần)

## Best Practices

1. ✅ **Không commit file vào Git** - đã có trong `.gitignore`
2. ✅ **Chỉ sử dụng trong development/testing** - dùng secret management trong production
3. ✅ **Rotate keys định kỳ** - ít nhất mỗi 6 tháng
4. ✅ **Giới hạn quyền** - chỉ cấp quyền cần thiết
5. ✅ **Giám sát sử dụng** - theo dõi trong Firebase Console
6. ✅ **Backup an toàn** - lưu file ở nơi an toàn (encrypted) nếu cần backup

## Tham Khảo

- [Firebase Admin SDK Setup](https://firebase.google.com/docs/admin/setup)
- [Service Accounts Documentation](https://firebase.google.com/docs/projects/service-accounts)
- [FCM Test Script](../../examples/test_fcm.go)
- [FCM Package README](../../pkg/fcm/README.md)
