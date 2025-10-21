# Hướng Dẫn Sử Dụng Swagger UI

## Giới Thiệu

Swagger UI cung cấp giao diện đồ họa để xem và test API một cách dễ dàng.

## Truy Cập Swagger UI

Sau khi khởi động server, bạn có thể truy cập Swagger UI tại:

```
http://localhost:3000/swagger
```

## Các Endpoint Swagger

| URL             | Mô tả                             |
| --------------- | --------------------------------- |
| `/swagger`      | Giao diện Swagger UI              |
| `/swagger.json` | File OpenAPI specification (JSON) |
| `/docs/*`       | Thư mục docs (các file static)    |

## Cách Test API Với Swagger UI

### 1. Mở Swagger UI

Truy cập `http://localhost:3000/swagger` trên trình duyệt

### 2. Xem Danh Sách API

- Swagger UI hiển thị tất cả endpoints được nhóm theo tags
- Click vào endpoint để xem chi tiết

### 3. Test API

1. Click vào endpoint muốn test
2. Click nút **"Try it out"**
3. Nhập parameters/body nếu cần
4. Click **"Execute"**
5. Xem response ở phía dưới

## Ví Dụ Test API

### Test GET /api/v1/users

1. Mở endpoint `GET /api/v1/users`
2. Click "Try it out"
3. Click "Execute"
4. Xem danh sách users trong response

### Test POST /api/v1/users

1. Mở endpoint `POST /api/v1/users`
2. Click "Try it out"
3. Nhập JSON body:

```json
{
  "name": "Nguyễn Văn A",
  "email": "nguyenvana@example.com"
}
```

4. Click "Execute"
5. Xem user mới được tạo trong response

### Test GET /api/v1/users/{id}

1. Mở endpoint `GET /api/v1/users/{id}`
2. Click "Try it out"
3. Nhập ID (copy từ response trước đó)
4. Click "Execute"
5. Xem thông tin user trong response

### Test PUT /api/v1/users/{id}

1. Mở endpoint `PUT /api/v1/users/{id}`
2. Click "Try it out"
3. Nhập ID
4. Nhập JSON body với thông tin mới:

```json
{
  "name": "Nguyễn Văn B",
  "email": "nguyenvanb@example.com"
}
```

5. Click "Execute"
6. Xem user đã được cập nhật

### Test DELETE /api/v1/users/{id}

1. Mở endpoint `DELETE /api/v1/users/{id}`
2. Click "Try it out"
3. Nhập ID
4. Click "Execute"
5. Response 204 nghĩa là xóa thành công

## Cấu Trúc File Swagger

### swagger.json

File này chứa OpenAPI specification theo chuẩn OpenAPI 3.0:

- **info**: Thông tin về API
- **servers**: Danh sách servers
- **paths**: Định nghĩa tất cả endpoints
- **components/schemas**: Định nghĩa data models
- **tags**: Nhóm các endpoints

### swagger.html

File HTML sử dụng Swagger UI library để hiển thị documentation.

## Cập Nhật Swagger Documentation

Khi thêm API mới, cập nhật file `docs/swagger.json`:

### Thêm Endpoint Mới

```json
"/api/v1/orders": {
  "get": {
    "summary": "Lấy danh sách orders",
    "tags": ["Orders"],
    "responses": {
      "200": {
        "description": "Danh sách orders",
        "content": {
          "application/json": {
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/components/schemas/Order"
              }
            }
          }
        }
      }
    }
  }
}
```

### Thêm Schema Mới

```json
"components": {
  "schemas": {
    "Order": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "example": "550e8400-e29b-41d4-a716-446655440000"
        },
        "user_id": {
          "type": "string"
        },
        "total": {
          "type": "number"
        }
      }
    }
  }
}
```

### Thêm Tag Mới

```json
"tags": [
  {
    "name": "Orders",
    "description": "Quản lý đơn hàng"
  }
]
```

## Sử Dụng Swagger Codegen

Bạn có thể sử dụng file `swagger.json` để generate client code:

```bash
# Generate TypeScript client
npx @openapitools/openapi-generator-cli generate \
  -i http://localhost:3000/swagger.json \
  -g typescript-axios \
  -o ./client

# Generate Go client
openapi-generator-cli generate \
  -i http://localhost:3000/swagger.json \
  -g go \
  -o ./go-client
```

## Tips

1. **Validation**: Swagger UI tự động validate input theo schema
2. **Examples**: Thêm examples trong swagger.json để dễ test
3. **Authentication**: Nếu API cần auth, thêm security schemes vào swagger.json
4. **Response Examples**: Thêm response examples để người dùng hiểu rõ hơn

## Troubleshooting

### Swagger UI không load được swagger.json

- Kiểm tra server có đang chạy không
- Kiểm tra đường dẫn file swagger.json
- Xem console browser để debug

### CORS errors

- Thêm CORS middleware nếu cần
- Đảm bảo server config đúng

### Swagger.json không hợp lệ

- Validate tại: https://editor.swagger.io/
- Kiểm tra JSON syntax
- Đảm bảo tuân thủ OpenAPI 3.0 spec
