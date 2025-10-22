# API Examples với Multi-language Support

Ví dụ sử dụng API với hỗ trợ đa ngôn ngữ.

## Language Detection

API tự động detect ngôn ngữ theo thứ tự:

1. Query parameter `lang`
2. Header `Accept-Language`
3. Default: `en`

## 1. Get All Users

### Request (English)

```bash
curl http://localhost:3000/api/v1/users
# hoặc
curl http://localhost:3000/api/v1/users?lang=en
```

### Response (English)

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Operation successful",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

### Request (Vietnamese)

```bash
curl http://localhost:3000/api/v1/users?lang=vi
# hoặc
curl -H "Accept-Language: vi" http://localhost:3000/api/v1/users
```

### Response (Vietnamese)

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Thành công",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

## 2. Get User By ID

### Request (English)

```bash
curl http://localhost:3000/api/v1/users/550e8400-e29b-41d4-a716-446655440001?lang=en
```

### Response (Success)

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Operation successful",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "John Doe",
    "email": "john@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Response (Not Found - English)

```json
{
  "success": false,
  "code": "USER_NOT_FOUND",
  "message": "User not found"
}
```

### Response (Not Found - Vietnamese)

```bash
curl http://localhost:3000/api/v1/users/invalid-id?lang=vi
```

```json
{
  "success": false,
  "code": "USER_NOT_FOUND",
  "message": "Không tìm thấy người dùng"
}
```

## 3. Create User

### Request (English)

```bash
curl -X POST http://localhost:3000/api/v1/users?lang=en \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Smith",
    "email": "jane@example.com"
  }'
```

### Response (Success - 201 Created)

```json
{
  "success": true,
  "code": "CREATED",
  "message": "Created successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "name": "Jane Smith",
    "email": "jane@example.com",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Request (Vietnamese)

```bash
curl -X POST http://localhost:3000/api/v1/users?lang=vi \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Nguyễn Văn A",
    "email": "nguyenvana@example.com"
  }'
```

### Response (Success - Vietnamese)

```json
{
  "success": true,
  "code": "CREATED",
  "message": "Tạo mới thành công",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440003",
    "name": "Nguyễn Văn A",
    "email": "nguyenvana@example.com",
    "created_at": "2024-01-15T10:35:00Z",
    "updated_at": "2024-01-15T10:35:00Z"
  }
}
```

### Response (Invalid Input - English)

```bash
curl -X POST http://localhost:3000/api/v1/users?lang=en \
  -H "Content-Type: application/json" \
  -d 'invalid json'
```

```json
{
  "success": false,
  "code": "INVALID_INPUT",
  "message": "Invalid input provided"
}
```

### Response (Invalid Input - Vietnamese)

```json
{
  "success": false,
  "code": "INVALID_INPUT",
  "message": "Dữ liệu đầu vào không hợp lệ"
}
```

## 4. Update User

### Request (English)

```bash
curl -X PUT http://localhost:3000/api/v1/users/550e8400-e29b-41d4-a716-446655440001?lang=en \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe Updated",
    "email": "john.updated@example.com"
  }'
```

### Response (Success)

```json
{
  "success": true,
  "code": "UPDATED",
  "message": "Updated successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "John Doe Updated",
    "email": "john.updated@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}
```

### Request (Vietnamese)

```bash
curl -X PUT http://localhost:3000/api/v1/users/550e8400-e29b-41d4-a716-446655440001?lang=vi \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Nguyễn Văn A Cập Nhật",
    "email": "updated@example.com"
  }'
```

### Response (Success - Vietnamese)

```json
{
  "success": true,
  "code": "UPDATED",
  "message": "Cập nhật thành công",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Nguyễn Văn A Cập Nhật",
    "email": "updated@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T11:00:00Z"
  }
}
```

## 5. Delete User

### Request (English)

```bash
curl -X DELETE http://localhost:3000/api/v1/users/550e8400-e29b-41d4-a716-446655440001?lang=en
```

### Response (Success)

```json
{
  "success": true,
  "code": "DELETED",
  "message": "Deleted successfully",
  "data": null
}
```

### Request (Vietnamese)

```bash
curl -X DELETE http://localhost:3000/api/v1/users/550e8400-e29b-41d4-a716-446655440001?lang=vi
```

### Response (Success - Vietnamese)

```json
{
  "success": true,
  "code": "DELETED",
  "message": "Xóa thành công",
  "data": null
}
```

## 6. Using Accept-Language Header

Thay vì dùng query parameter, có thể dùng header:

### Request

```bash
curl http://localhost:3000/api/v1/users \
  -H "Accept-Language: vi,en-US;q=0.9,en;q=0.8"
```

### Response

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Thành công",
  "data": [...]
}
```

## 7. Error Examples

### Internal Server Error (English)

```json
{
  "success": false,
  "code": "INTERNAL_SERVER_ERROR",
  "message": "Internal server error"
}
```

### Internal Server Error (Vietnamese)

```json
{
  "success": false,
  "code": "INTERNAL_SERVER_ERROR",
  "message": "Lỗi máy chủ"
}
```

## Testing Tips

### 1. Test với curl

```bash
# English
curl "http://localhost:3000/api/v1/users?lang=en"

# Vietnamese
curl "http://localhost:3000/api/v1/users?lang=vi"

# Với header
curl -H "Accept-Language: vi" "http://localhost:3000/api/v1/users"
```

### 2. Test với Postman

- Thêm query parameter `lang=vi` hoặc `lang=en`
- Hoặc thêm header `Accept-Language: vi`

### 3. Test trong browser

```
http://localhost:3000/api/v1/users?lang=vi
http://localhost:3000/api/v1/users?lang=en
```

## Response Format Summary

### Success Response

```json
{
  "success": true,
  "code": "CODE",
  "message": "Translated message",
  "data": {...}
}
```

### Error Response

```json
{
  "success": false,
  "code": "ERROR_CODE",
  "message": "Translated error message",
  "errors": {...} // optional
}
```

### Optional Status Code

Các hàm response đều cho phép truyền custom HTTP status code (optional):

```go
// Sử dụng default status code
response.Success(w, lang, response.CodeSuccess, data)

// Truyền custom status code
response.Success(w, lang, response.CodeSuccess, data, 200)

// Error với custom status code
response.Error(w, lang, response.CodeCustomError, nil, 400)
```

## Supported Languages

- `en` - English (default)
- `vi` - Tiếng Việt

## Common Response Codes

| Code                  | HTTP Status | English Message        | Vietnamese Message           |
| --------------------- | ----------- | ---------------------- | ---------------------------- |
| SUCCESS               | 200         | Operation successful   | Thành công                   |
| CREATED               | 201         | Created successfully   | Tạo mới thành công           |
| UPDATED               | 200         | Updated successfully   | Cập nhật thành công          |
| DELETED               | 200         | Deleted successfully   | Xóa thành công               |
| INVALID_INPUT         | 400         | Invalid input provided | Dữ liệu đầu vào không hợp lệ |
| USER_NOT_FOUND        | 404         | User not found         | Không tìm thấy người dùng    |
| INTERNAL_SERVER_ERROR | 500         | Internal server error  | Lỗi máy chủ                  |
