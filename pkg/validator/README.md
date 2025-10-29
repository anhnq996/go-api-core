# Validator Package

Package validation tự động sử dụng struct tags với `go-playground/validator/v10`.

## Features

- ✅ Auto parse JSON request body
- ✅ Auto validate với struct tags
- ✅ Auto response errors với field details
- ✅ Custom validators (phone, strongpassword)
- ✅ Sử dụng JSON tags cho field names
- ✅ Support 30+ validation rules có sẵn

## Quick Start

### 1. Define Request Struct

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6"`
}
```

### 2. Validate trong Controller

```go
func Login(w http.ResponseWriter, r *http.Request) {
    var input LoginRequest

    // Tự động parse JSON + validate + response errors
    if !validator.ValidateAndRespond(w, r, &input) {
        return // Validation failed, đã tự động response
    }

    // input đã được validate, sử dụng an toàn
    // ... business logic
}
```

### Response khi validation failed:

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

### Required Rules

```go
type Request struct {
    Name  string `validate:"required"`              // Bắt buộc
    Email string `validate:"required,email"`         // Bắt buộc + email
    Age   int    `validate:"required,gte=18"`       // Bắt buộc + >= 18
}
```

### String Rules

```go
type Request struct {
    Username string `validate:"required,min=3,max=20"`     // Độ dài 3-20
    Code     string `validate:"required,len=6"`            // Đúng 6 ký tự
    Slug     string `validate:"required,alpha"`            // Chỉ chữ
    Token    string `validate:"required,alphanum"`         // Chữ + số
    Phone    string `validate:"required,numeric"`          // Chỉ số
    URL      string `validate:"required,url"`              // URL format
    UUID     string `validate:"required,uuid"`             // UUID format
}
```

### Number Rules

```go
type Request struct {
    Age      int     `validate:"required,gte=18,lte=100"`  // 18 <= age <= 100
    Price    float64 `validate:"required,gt=0"`            // > 0
    Quantity int     `validate:"required,min=1,max=1000"`  // 1-1000
}
```

### Email & URLs

```go
type Request struct {
    Email    string `validate:"required,email"`
    Website  string `validate:"required,url"`
    Avatar   string `validate:"omitempty,url"`  // Optional nhưng phải là URL
}
```

### Custom Validators

```go
type Request struct {
    Phone    string `validate:"required,phone"`           // Vietnamese phone
    Password string `validate:"required,strongpassword"`  // Strong password
}
```

#### Phone Validator

- 10 chữ số
- Bắt đầu bằng 0
- Ví dụ: 0123456789

#### StrongPassword Validator

- Ít nhất 8 ký tự
- Có chữ hoa
- Có chữ thường
- Có số
- Có ký tự đặc biệt (!@#$%^&\*)

### Conditional Validation

```go
type Request struct {
    Type     string `validate:"required,oneof=user admin"` // Chỉ "user" hoặc "admin"
    Email    string `validate:"required_if=Type user"`     // Required nếu Type=user
    Optional string `validate:"omitempty,email"`           // Optional, nhưng nếu có phải là email
}
```

### Cross-Field Validation

```go
type Request struct {
    Password        string `validate:"required,min=8"`
    ConfirmPassword string `validate:"required,eqfield=Password"` // Phải giống Password

    StartDate time.Time `validate:"required"`
    EndDate   time.Time `validate:"required,gtfield=StartDate"` // Phải > StartDate
}
```

### Nested Structs

```go
type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    ZipCode string `validate:"required,numeric,len=5"`
}

type UserRequest struct {
    Name    string  `validate:"required"`
    Email   string  `validate:"required,email"`
    Address Address `validate:"required"` // Validate nested struct
}
```

### Slices & Arrays

```go
type Request struct {
    Tags     []string `validate:"required,min=1,max=10"`           // 1-10 items
    Emails   []string `validate:"required,dive,email"`            // Mỗi item phải là email
    Numbers  []int    `validate:"required,dive,gte=0,lte=100"`    // Mỗi item 0-100
}
```

## Complete Example

### Request Definition

```go
// internal/app/user/request.go
package user

type CreateUserRequest struct {
    Name     string  `json:"name" validate:"required,min=2,max=100"`
    Email    string  `json:"email" validate:"required,email"`
    Password string  `json:"password" validate:"required,strongpassword"`
    Phone    string  `json:"phone" validate:"required,phone"`
    Age      int     `json:"age" validate:"required,gte=18,lte=150"`
    Avatar   *string `json:"avatar" validate:"omitempty,url"`
    Role     string  `json:"role" validate:"required,oneof=user admin moderator"`
}

type UpdateUserRequest struct {
    Name   string  `json:"name" validate:"omitempty,min=2,max=100"`
    Phone  string  `json:"phone" validate:"omitempty,phone"`
    Avatar *string `json:"avatar" validate:"omitempty,url"`
}
```

### Controller

```go
// internal/app/user/controller.go
package user

import (
    "api-core/pkg/validator"
    "api-core/pkg/response"
)

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    var input CreateUserRequest

    // Tự động validate
    if !validator.ValidateAndRespond(w, r, &input) {
        return // Validation failed, response sent
    }

    // Validation passed, continue
    user, err := h.service.Create(r.Context(), input)
    if err != nil {
        response.InternalServerError(w, lang, response.CodeInternalServerError)
        return
    }

    response.Created(w, lang, response.CodeCreated, user)
}
```

## Manual Validation

Nếu không muốn auto response:

```go
import "api-core/pkg/validator"

func Handler(w http.ResponseWriter, r *http.Request) {
    var input LoginRequest

    // Parse JSON manually
    json.NewDecoder(r.Body).Decode(&input)

    // Validate manually
    if err := validator.Validate(&input); err != nil {
        // Parse errors
        validationErrors := validator.ParseValidationErrors(err)

        // Custom response
        response.ValidationError(w, lang, response.CodeValidationFailed, validationErrors)
        return
    }

    // Continue...
}
```

## Available Validation Tags

### Basic

- `required` - Bắt buộc
- `omitempty` - Optional (skip validation nếu empty)

### String

- `min=n` - Độ dài tối thiểu
- `max=n` - Độ dài tối đa
- `len=n` - Độ dài chính xác
- `eq=value` - Bằng giá trị
- `ne=value` - Không bằng giá trị
- `oneof=val1 val2` - Một trong các giá trị

### String Format

- `email` - Email format
- `url` - URL format
- `uri` - URI format
- `uuid` - UUID format
- `alpha` - Chỉ chữ cái
- `alphanum` - Chữ cái + số
- `numeric` - Chỉ số
- `lowercase` - Chỉ chữ thường
- `uppercase` - Chỉ chữ hoa

### Number

- `gt=n` - Lớn hơn
- `gte=n` - Lớn hơn hoặc bằng
- `lt=n` - Nhỏ hơn
- `lte=n` - Nhỏ hơn hoặc bằng
- `eq=n` - Bằng
- `ne=n` - Không bằng

### Cross-Field

- `eqfield=Field` - Bằng field khác
- `nefield=Field` - Không bằng field khác
- `gtfield=Field` - Lớn hơn field khác
- `ltfield=Field` - Nhỏ hơn field khác

### Custom (Project-specific)

- `phone` - Vietnamese phone number (0xxxxxxxxx)
- `strongpassword` - Strong password (8+ chars, upper, lower, number, special)

### Collections

- `dive` - Validate từng item trong slice/array
- `unique` - Items phải unique

## Error Messages

Package tự động generate error messages:

| Tag              | Example Message                                                            |
| ---------------- | -------------------------------------------------------------------------- |
| required         | "email is required"                                                        |
| email            | "email must be a valid email address"                                      |
| min=8            | "password must be at least 8 characters"                                   |
| max=100          | "name must not exceed 100 characters"                                      |
| gte=18           | "age must be greater than or equal to 18"                                  |
| eqfield=Password | "confirm_password must be equal to Password"                               |
| strongpassword   | "password must contain uppercase, lowercase, number and special character" |

## Testing

```go
func TestValidation(t *testing.T) {
    input := LoginRequest{
        Email:    "invalid-email",
        Password: "123", // Too short
    }

    err := validator.Validate(&input)
    assert.Error(t, err)

    errors := validator.ParseValidationErrors(err)
    assert.Len(t, errors, 2)
    assert.Equal(t, "email", errors[0].Field)
    assert.Equal(t, "password", errors[1].Field)
}
```

## Add Custom Validator

```go
// pkg/validator/validator.go

func registerCustomValidators() {
    validate := validator.GetValidator()

    // Thêm custom validator mới
    validate.RegisterValidation("customrule", func(fl validator.FieldLevel) bool {
        value := fl.Field().String()
        // Your validation logic
        return true
    })
}
```

Usage:

```go
type Request struct {
    CustomField string `validate:"required,customrule"`
}
```

## Best Practices

### 1. Sử dụng struct tags

```go
// ✅ Good
type Request struct {
    Email string `json:"email" validate:"required,email"`
}

// ❌ Bad - Manual validation
func validate(input Request) error {
    if input.Email == "" {
        return errors.New("email required")
    }
    // ...
}
```

### 2. Luôn dùng ValidateAndRespond

```go
// ✅ Good - 1 dòng code
if !validator.ValidateAndRespond(w, r, &input) {
    return
}

// ❌ Bad - Nhiều code, dễ sai
var input Request
json.NewDecoder(r.Body).Decode(&input)
if err := validator.Validate(&input); err != nil {
    errors := validator.ParseValidationErrors(err)
    response.ValidationError(w, lang, code, errors)
    return
}
```

### 3. Group validation rules

```go
// ✅ Good
type CreateUserRequest struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,strongpassword"`
}

type UpdateUserRequest struct {
    Email string `validate:"omitempty,email"`  // Optional
}
```

### 4. Tái sử dụng validation groups

```go
// Common fields
type BaseUserFields struct {
    Name  string `validate:"required,min=2,max=100"`
    Email string `validate:"required,email"`
}

// Extend
type CreateUserRequest struct {
    BaseUserFields
    Password string `validate:"required,strongpassword"`
}

type UpdateUserRequest struct {
    Name  string `validate:"omitempty,min=2,max=100"`
    Email string `validate:"omitempty,email"`
}
```

## Resources

- [Validator Documentation](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [Validation Tags](https://github.com/go-playground/validator#baked-in-validations)
