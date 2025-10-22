# Validation Guide

Hướng dẫn sử dụng validation tự động trong ApiCore.

## Tổng quan

Validation package sử dụng struct tags để khai báo rules, tự động:

- ✅ Parse JSON request body
- ✅ Validate theo rules
- ✅ Response errors với field details
- ✅ Support 30+ validation rules
- ✅ Custom validators (phone, strongpassword)

## Quick Start

### 1. Định nghĩa Request Struct

```go
// internal/app/auth/request.go
package auth

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}
```

**Validation rules:**

- `email`: Bắt buộc + phải là email hợp lệ
- `password`: Bắt buộc + tối thiểu 6 ký tự

### 2. Sử dụng trong Controller

```go
// internal/app/auth/controller.go
import "anhnq/api-core/pkg/validator"

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var input LoginRequest

    // 1 dòng code: parse JSON + validate + auto response errors
    if !validator.ValidateAndRespond(w, r, &input) {
        return // Validation failed, response đã gửi
    }

    // Validation passed, sử dụng input an toàn
    result, err := h.service.Login(r.Context(), input.Email, input.Password)
    // ...
}
```

### 3. Response khi Validation Failed

**Request:**

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"invalid","password":"123"}'
```

**Response (422):**

```json
{
  "success": false,
  "code": "VALIDATION_FAILED",
  "message": "Xác thực dữ liệu thất bại",
  "errors": [
    {
      "field": "email",
      "message": "email must be a valid email address"
    },
    {
      "field": "password",
      "message": "password must be at least 6 characters"
    }
  ]
}
```

## Validation Rules

### Common Rules

```go
type UserRequest struct {
    // Required
    Name  string `validate:"required"`
    Email string `validate:"required,email"`

    // Optional (omitempty = skip validation if empty)
    Phone  string `validate:"omitempty,phone"`
    Avatar string `validate:"omitempty,url"`

    // Length constraints
    Username string `validate:"required,min=3,max=20"`
    Code     string `validate:"required,len=6"` // Exactly 6 chars

    // Number constraints
    Age      int     `validate:"required,gte=18,lte=150"` // 18 <= age <= 150
    Price    float64 `validate:"required,gt=0"`           // > 0
    Quantity int     `validate:"required,min=1,max=100"`  // 1-100

    // Format validation
    Website string `validate:"required,url"`
    UUID    string `validate:"required,uuid"`

    // Enum validation
    Role   string `validate:"required,oneof=user admin moderator"`
    Status string `validate:"oneof=active inactive pending"`
}
```

### Password Validation

```go
type RegisterRequest struct {
    Password        string `validate:"required,strongpassword"`
    ConfirmPassword string `validate:"required,eqfield=Password"`
}
```

**StrongPassword requirements:**

- Ít nhất 8 ký tự
- Có chữ hoa (A-Z)
- Có chữ thường (a-z)
- Có số (0-9)
- Có ký tự đặc biệt (!@#$%^&\*)

Examples:

- ✅ `Password123!`
- ✅ `MyPass@2024`
- ❌ `password` (no upper, number, special)
- ❌ `Pass1!` (too short)

### Phone Validation

```go
type Request struct {
    Phone string `validate:"required,phone"`
}
```

**Vietnamese phone format:**

- 10 chữ số
- Bắt đầu bằng 0
- ✅ `0123456789`
- ✅ `0987654321`
- ❌ `123456789` (missing 0)
- ❌ `01234567890` (too long)

### Cross-Field Validation

```go
type ChangePasswordRequest struct {
    NewPassword     string `validate:"required,min=8"`
    ConfirmPassword string `validate:"required,eqfield=NewPassword"`
}

type DateRangeRequest struct {
    StartDate time.Time `validate:"required"`
    EndDate   time.Time `validate:"required,gtfield=StartDate"`
}
```

### Nested Structs

```go
type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    ZipCode string `validate:"required,len=5,numeric"`
}

type UserRequest struct {
    Name    string  `validate:"required"`
    Address Address `validate:"required"` // Validate nested
}
```

### Arrays/Slices

```go
type Request struct {
    Tags    []string `validate:"required,min=1,max=5"`          // 1-5 items
    Emails  []string `validate:"required,dive,email"`          // Each item = email
    Numbers []int    `validate:"required,dive,gte=0,lte=100"` // Each 0-100
}
```

## Real-World Examples

### Auth Module

```go
// internal/app/auth/request.go
package auth

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,strongpassword"`
}

type ChangePasswordRequest struct {
    CurrentPassword string `json:"current_password" validate:"required"`
    NewPassword     string `json:"new_password" validate:"required,strongpassword"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
```

### User Module

```go
// internal/app/user/request.go
package user

type CreateUserRequest struct {
    Name   string  `json:"name" validate:"required,min=2,max=100"`
    Email  string  `json:"email" validate:"required,email"`
    Phone  string  `json:"phone" validate:"omitempty,phone"`
    RoleID *string `json:"role_id" validate:"omitempty,uuid"`
}

type UpdateUserRequest struct {
    Name   string  `json:"name" validate:"omitempty,min=2,max=100"`
    Email  string  `json:"email" validate:"omitempty,email"`
    Avatar *string `json:"avatar" validate:"omitempty,url"`
}
```

### Product Module (Example)

```go
type CreateProductRequest struct {
    Name        string   `json:"name" validate:"required,min=3,max=200"`
    Description string   `json:"description" validate:"required,min=10"`
    Price       float64  `json:"price" validate:"required,gt=0"`
    Stock       int      `json:"stock" validate:"required,gte=0"`
    Category    string   `json:"category" validate:"required,oneof=electronics clothing food"`
    Tags        []string `json:"tags" validate:"omitempty,max=10,dive,min=2,max=20"`
    Images      []string `json:"images" validate:"omitempty,max=5,dive,url"`
}
```

## Manual Validation

Nếu cần kiểm soát response tùy chỉnh:

```go
import "anhnq/api-core/pkg/validator"

func Handler(w http.ResponseWriter, r *http.Request) {
    var input LoginRequest

    // Manual parse
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        // Handle parse error
        return
    }

    // Manual validate
    if err := validator.Validate(&input); err != nil {
        errors := validator.ParseValidationErrors(err)

        // Custom response
        response.ValidationError(w, lang, response.CodeValidationFailed, errors)
        return
    }

    // Continue...
}
```

## Error Messages

Package tự động generate error messages tiếng Anh:

| Rule             | Message Template                                                          |
| ---------------- | ------------------------------------------------------------------------- |
| required         | "{field} is required"                                                     |
| email            | "{field} must be a valid email address"                                   |
| min=8            | "{field} must be at least 8 characters"                                   |
| max=100          | "{field} must not exceed 100 characters"                                  |
| gte=18           | "{field} must be greater than or equal to 18"                             |
| eqfield=Password | "{field} must be equal to Password"                                       |
| phone            | "{field} must be a valid phone number"                                    |
| strongpassword   | "{field} must contain uppercase, lowercase, number and special character" |

## Add Custom Validator

```go
// pkg/validator/validator.go

func registerCustomValidators() {
    // ... existing validators

    // Thêm custom validator
    validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
        username := fl.Field().String()
        // Check: 3-20 chars, alphanumeric, underscore, dash
        matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{3,20}$`, username)
        return matched
    })
}
```

**Usage:**

```go
type Request struct {
    Username string `validate:"required,username"`
}
```

## Testing

```go
func TestLoginValidation(t *testing.T) {
    // Valid request
    input := LoginRequest{
        Email:    "user@example.com",
        Password: "password123",
    }
    err := validator.Validate(&input)
    assert.NoError(t, err)

    // Invalid email
    input.Email = "invalid-email"
    err = validator.Validate(&input)
    assert.Error(t, err)

    errors := validator.ParseValidationErrors(err)
    assert.Equal(t, "email", errors[0].Field)
}
```

## Best Practices

### 1. Luôn dùng ValidateAndRespond

```go
// ✅ Good - 1 line, tự động mọi thứ
if !validator.ValidateAndRespond(w, r, &input) {
    return
}

// ❌ Bad - Manual, nhiều code
var input Request
json.NewDecoder(r.Body).Decode(&input)
if input.Email == "" {
    response.BadRequest(...)
    return
}
if !utils.IsEmail(input.Email) {
    response.BadRequest(...)
    return
}
```

### 2. Tách request structs vào file riêng

```go
// ✅ Good structure
internal/app/auth/
├── controller.go
├── service.go
├── request.go     ← Request validation structs
└── route.go
```

### 3. Sử dụng omitempty cho optional fields

```go
// ✅ Good
type UpdateRequest struct {
    Name  string `validate:"omitempty,min=2"` // Optional, nhưng nếu có phải >= 2 chars
    Email string `validate:"omitempty,email"` // Optional, nhưng nếu có phải là email
}

// ❌ Bad
type UpdateRequest struct {
    Name  string `validate:"required,min=2"` // Luôn bắt buộc
}
```

### 4. Validation rules rõ ràng

```go
// ✅ Good - Rules rõ ràng
type Request struct {
    Age int `validate:"required,gte=18,lte=150"`
}

// ❌ Unclear
type Request struct {
    Age int `validate:"required,min=18"` // min cho string, gte cho number
}
```

## Integration Example

### Complete Flow

```go
// 1. Define request
type CreateProductRequest struct {
    Name  string  `json:"name" validate:"required,min=3,max=200"`
    Price float64 `json:"price" validate:"required,gt=0"`
}

// 2. Controller
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateProductRequest

    if !validator.ValidateAndRespond(w, r, &input) {
        return
    }

    product := h.service.Create(input)
    response.Created(w, lang, response.CodeCreated, product)
}

// 3. Test
curl -X POST /api/v1/products \
  -d '{"name":"A","price":-10}' # Invalid

# Response:
# {
#   "success": false,
#   "errors": [
#     {"field":"name","message":"name must be at least 3 characters"},
#     {"field":"price","message":"price must be greater than 0"}
#   ]
# }
```

## Troubleshooting

### Problem: Validation không hoạt động

**Check:**

1. Struct có tag `validate:""` không?
2. Field phải exported (bắt đầu bằng chữ hoa)
3. JSON tag match với request body

### Problem: Custom validator không chạy

**Check:**

1. Đã register trong `registerCustomValidators()`?
2. Function signature đúng: `func(fl validator.FieldLevel) bool`

### Problem: Error message không đúng

**Check:**

1. JSON tag để hiển thị field name đẹp
2. Update `GetErrorMessage()` cho custom messages

## See Also

- [pkg/validator/README.md](../pkg/validator/README.md)
- [Validator Documentation](https://pkg.go.dev/github.com/go-playground/validator/v10)
