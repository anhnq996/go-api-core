# Response Package

Package chuẩn hóa format REST API response với hỗ trợ đa ngôn ngữ (i18n).

## Format Response

### Success Response

```json
{
  "status": "success",
  "code": "SUCCESS",
  "message": "Operation successful",
  "data": {
    "id": 123,
    "name": "John Doe"
  }
}
```

### Success Response With Pagination

```json
{
  "status": "success",
  "code": "SUCCESS",
  "message": "Operation successful",
  "data": [
    { "id": 1, "name": "User 1" },
    { "id": 2, "name": "User 2" }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

### Error Response

```json
{
  "status": "error",
  "code": "VALIDATION_FAILED",
  "message": "Validation failed",
  "errors": [
    {
      "field": "email",
      "message": "Email is required"
    },
    {
      "field": "password",
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

## Cài đặt

### 1. Khởi tạo i18n

Trong `main.go`:

```go
import (
    "api-core/pkg/i18n"
    "api-core/pkg/response"
)

func main() {
    // Khởi tạo i18n
    err := i18n.Init(i18n.Config{
        TranslationsDir: "translations",
        Languages:       []string{"en", "vi"},
        FallbackLang:    "en",
    })
    if err != nil {
        log.Fatal(err)
    }

    // ... setup router, etc.
}
```

### 2. Cấu trúc thư mục translations

```
translations/
  ├── en.json  # Tiếng Anh
  └── vi.json  # Tiếng Việt
```

## Sử dụng

### Success Response

```go
import (
    "api-core/pkg/response"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
    // Lấy ngôn ngữ từ request
    lang := response.GetLanguageFromRequest(r)

    user := map[string]interface{}{
        "id":    123,
        "name":  "John Doe",
        "email": "john@example.com",
    }

    // Sử dụng default status code (200)
    response.Success(w, lang, response.CodeSuccess, user)

    // Hoặc truyền custom status code
    // response.Success(w, lang, response.CodeSuccess, user, 200)
}
```

### Created Response (201)

```go
func CreateUser(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    newUser := map[string]interface{}{
        "id":    456,
        "email": "newuser@example.com",
    }

    response.Created(w, lang, response.CodeCreated, newUser)
}
```

### Success With Pagination

```go
func ListUsers(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    users := []User{} // Lấy từ database

    meta := &response.Meta{
        Page:       1,
        PerPage:    10,
        Total:      100,
        TotalPages: 10,
    }

    response.SuccessWithMeta(w, lang, response.CodeSuccess, users, meta)
}
```

### Error Responses

#### Validation Error (422)

```go
func CreateUser(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    errors := []response.ErrorDetail{
        {Field: "email", Message: "Email is required"},
        {Field: "password", Message: "Password must be at least 8 characters"},
    }

    response.ValidationError(w, lang, response.CodeValidationFailed, errors)
}
```

#### Not Found (404)

```go
func GetUser(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    // User không tồn tại
    response.NotFound(w, lang, response.CodeUserNotFound)
}
```

#### Unauthorized (401)

```go
func Login(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    // Sai username/password
    response.Unauthorized(w, lang, response.CodeInvalidCredentials)
}
```

#### Forbidden (403)

```go
func DeleteUser(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    // Không có quyền
    response.Forbidden(w, lang, response.CodePermissionDenied)
}
```

#### Conflict (409)

```go
func CreateUser(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    // Email đã tồn tại
    response.Conflict(w, lang, response.CodeEmailAlreadyExists)
}
```

#### Internal Server Error (500)

```go
func GetUser(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    // Lỗi server
    response.InternalServerError(w, lang, response.CodeInternalServerError)
}
```

#### Custom Error

```go
func ProcessPayment(w http.ResponseWriter, r *http.Request) {
    lang := response.GetLanguageFromRequest(r)

    errorDetails := map[string]interface{}{
        "current_balance": 1000,
        "required_amount": 5000,
    }

    response.Error(
        w,
        http.StatusBadRequest,
        lang,
        response.CodeInsufficientBalance,
        errorDetails,
    )
}
```

## Phát hiện ngôn ngữ

Package tự động phát hiện ngôn ngữ theo thứ tự ưu tiên:

1. **Query parameter** `lang`:

   ```
   GET /api/users?lang=vi
   ```

2. **Header** `Accept-Language`:

   ```
   Accept-Language: vi,en-US;q=0.9,en;q=0.8
   ```

3. **Default**: `en` (English)

### Ví dụ

```go
func GetUser(w http.ResponseWriter, r *http.Request) {
    // Tự động detect: query param -> header -> default
    lang := response.GetLanguageFromRequest(r)

    // Hoặc force một ngôn ngữ cụ thể
    lang = "vi"

    response.Success(w, lang, response.CodeSuccess, userData)
}
```

## Response Codes

Package cung cấp các response codes chuẩn:

### Success Codes

- `CodeSuccess` - Thành công chung
- `CodeCreated` - Tạo mới thành công
- `CodeUpdated` - Cập nhật thành công
- `CodeDeleted` - Xóa thành công
- `CodeNoContent` - Không có nội dung

### Client Error Codes

- `CodeBadRequest` - Request không hợp lệ
- `CodeInvalidInput` - Input không hợp lệ
- `CodeValidationFailed` - Validation thất bại
- `CodeUnauthorized` - Chưa xác thực
- `CodeForbidden` - Không có quyền
- `CodeNotFound` - Không tìm thấy
- `CodeConflict` - Xung đột
- `CodeDuplicateEntry` - Dữ liệu trùng lặp

### Authentication Codes

- `CodeInvalidCredentials` - Sai username/password
- `CodeTokenExpired` - Token hết hạn
- `CodeTokenInvalid` - Token không hợp lệ
- `CodeTokenMissing` - Thiếu token
- `CodePermissionDenied` - Không có quyền

### Server Error Codes

- `CodeInternalServerError` - Lỗi server
- `CodeServiceUnavailable` - Service không khả dụng
- `CodeDatabaseError` - Lỗi database
- `CodeCacheError` - Lỗi cache

### Business Logic Codes

- `CodeInsufficientBalance` - Số dư không đủ
- `CodeOperationFailed` - Thao tác thất bại
- `CodeInvalidOperation` - Thao tác không hợp lệ
- `CodeLimitExceeded` - Vượt quá giới hạn

### User Codes

- `CodeUserNotFound` - Không tìm thấy user
- `CodeUserAlreadyExists` - User đã tồn tại
- `CodeEmailAlreadyExists` - Email đã tồn tại
- `CodePhoneAlreadyExists` - SĐT đã tồn tại

Xem đầy đủ trong `pkg/response/codes.go`

## Thêm Response Code Mới

### 1. Thêm code vào `pkg/response/codes.go`

```go
const (
    // ... existing codes
    CodeCustomError = "CUSTOM_ERROR"
)
```

### 2. Thêm translations vào `translations/en.json`

```json
{
  "CUSTOM_ERROR": "This is a custom error message"
}
```

### 3. Thêm translations vào `translations/vi.json`

```json
{
  "CUSTOM_ERROR": "Đây là thông báo lỗi tùy chỉnh"
}
```

### 4. Sử dụng

```go
response.Error(w, 400, lang, response.CodeCustomError, nil)
```

## Thêm Ngôn Ngữ Mới

### 1. Tạo file translation mới

Ví dụ `translations/ja.json` cho tiếng Nhật:

```json
{
  "SUCCESS": "成功しました",
  "CREATED": "作成されました",
  ...
}
```

### 2. Cập nhật config khi init

```go
i18n.Init(i18n.Config{
    TranslationsDir: "translations",
    Languages:       []string{"en", "vi", "ja"},
    FallbackLang:    "en",
})
```

## Message với Parameters

Bạn có thể dùng format string trong translations:

### translations/en.json

```json
{
  "USER_CREATED": "User %s created successfully",
  "ITEMS_FOUND": "Found %d items"
}
```

### translations/vi.json

```json
{
  "USER_CREATED": "Tạo người dùng %s thành công",
  "ITEMS_FOUND": "Tìm thấy %d mục"
}
```

### Sử dụng

```go
// Với 1 parameter
response.Success(w, lang, "USER_CREATED", userData, "john@example.com")

// Với nhiều parameters
response.Success(w, lang, "ITEMS_FOUND", items, 25)
```

## Best Practices

### 1. Luôn lấy language từ request

```go
lang := response.GetLanguageFromRequest(r)
```

### 2. Sử dụng response codes có sẵn

```go
// Good
response.NotFound(w, lang, response.CodeUserNotFound)

// Avoid
response.Error(w, 404, lang, "USER_NOT_FOUND", nil)
```

### 3. Validation errors nên chi tiết

```go
errors := []response.ErrorDetail{
    {Field: "email", Message: "Email must be valid"},
    {Field: "password", Message: "Password must be at least 8 characters"},
}
response.ValidationError(w, lang, response.CodeValidationFailed, errors)
```

### 4. Log errors trước khi trả response

```go
if err != nil {
    logger.ErrorWithErr(err, "Failed to get user")
    response.InternalServerError(w, lang, response.CodeInternalServerError)
    return
}
```

### 5. Không expose sensitive info trong error response

```go
// Bad
response.InternalServerError(w, lang, response.CodeDatabaseError)

// Good - Log internally, generic message to client
logger.ErrorWithErr(err, "Database connection failed: "+err.Error())
response.InternalServerError(w, lang, response.CodeInternalServerError)
```

## Testing

```go
func TestResponseFormat(t *testing.T) {
    w := httptest.NewRecorder()

    data := map[string]interface{}{"id": 123}
    response.Success(w, "en", response.CodeSuccess, data)

    var resp response.Response
    json.Unmarshal(w.Body.Bytes(), &resp)

    assert.Equal(t, "success", resp.Status)
    assert.Equal(t, "SUCCESS", resp.Code)
    assert.NotEmpty(t, resp.Message)
}
```
