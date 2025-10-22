# Authentication Module

Module authentication hoàn chỉnh với JWT và RBAC (Role-Based Access Control).

## Features

- ✅ Login với email & password
- ✅ Register user mới
- ✅ Logout (single device)
- ✅ Logout all devices
- ✅ Refresh token
- ✅ Get current user info
- ✅ Role-Based Access Control (RBAC)
- ✅ Permissions system
- ✅ Token blacklist
- ✅ Password hashing với bcrypt
- ✅ Multi-language response

## Database Schema

### Tables

1. **users** - User accounts

   - id (UUID, PK)
   - name
   - email (unique)
   - password (hashed)
   - avatar (optional)
   - role_id (FK to roles)
   - email_verified_at
   - is_active
   - last_login_at

2. **roles** - User roles

   - id (UUID, PK)
   - name (unique: admin, moderator, user)
   - display_name
   - description

3. **permissions** - System permissions

   - id (UUID, PK)
   - name (unique: users.view, users.create, etc.)
   - display_name
   - description
   - module

4. **role_has_permissions** - Role-Permission relationships
   - role_id (FK)
   - permission_id (FK)

## Default Seeded Data

### Roles

- **admin** - Full system access
- **moderator** - Content & user management
- **user** - Basic permissions

### Users

| Email                 | Password     | Role      |
| --------------------- | ------------ | --------- |
| admin@example.com     | Password123! | admin     |
| moderator@example.com | Password123! | moderator |
| user@example.com      | Password123! | user      |
| john@example.com      | Password123! | user      |
| jane@example.com      | Password123! | user      |

### Permissions

**Users Module:**

- users.view
- users.create
- users.update
- users.delete

**Roles Module:**

- roles.view
- roles.manage

**Permissions Module:**

- permissions.view
- permissions.manage

**Profile Module:**

- profile.view
- profile.update

## API Endpoints

### Public Routes

#### 1. Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "Password123!"
}
```

**Response (200):**

```json
{
  "success": true,
  "code": "LOGIN_SUCCESS",
  "message": "Đăng nhập thành công",
  "data": {
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Admin User",
      "email": "admin@example.com",
      "avatar": null,
      "role": {
        "id": "11111111-1111-1111-1111-111111111111",
        "name": "admin",
        "display_name": "Administrator"
      },
      "permissions": [
        "users.view",
        "users.create",
        "users.update",
        "users.delete",
        "roles.view",
        "roles.manage",
        "permissions.view",
        "permissions.manage",
        "profile.view",
        "profile.update"
      ]
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2024-01-15T11:00:00Z",
    "token_type": "Bearer"
  }
}
```

#### 2. Register

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "name": "New User",
  "email": "newuser@example.com",
  "password": "SecurePass123!"
}
```

**Response (201):**

```json
{
  "success": true,
  "code": "CREATED",
  "message": "Tạo mới thành công",
  "data": {
    "id": "...",
    "name": "New User",
    "email": "newuser@example.com",
    "is_active": true
  }
}
```

#### 3. Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Protected Routes

#### 4. Get Current User

```http
GET /api/v1/auth/me
Authorization: Bearer <access_token>
```

**Response (200):**

```json
{
  "success": true,
  "code": "SUCCESS",
  "message": "Thành công",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Admin User",
    "email": "admin@example.com",
    "avatar": null,
    "role": {
      "id": "11111111-1111-1111-1111-111111111111",
      "name": "admin",
      "display_name": "Administrator"
    },
    "permissions": [...]
  }
}
```

#### 5. Logout

```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

#### 6. Logout All Devices

```http
POST /api/v1/auth/logout-all
Authorization: Bearer <access_token>
```

## Setup & Run

### 1. Run Migrations

```bash
make migrate

# Hoặc
go run cmd/migrate/main.go up
```

### 2. Run Seeders

```bash
make seed

# Hoặc
go run cmd/migrate/main.go seed
```

### 3. Start Server

```bash
make watch
```

## Testing với curl

### Login

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "Password123!"
  }'
```

### Get Current User

```bash
# Save access_token từ login response
TOKEN="<access_token>"

curl http://localhost:3000/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

### Logout

```bash
curl -X POST http://localhost:3000/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

## Role-Based Permissions

### Admin Role

Có tất cả permissions:

- users.\* (view, create, update, delete)
- roles.\* (view, manage)
- permissions.\* (view, manage)
- profile.\* (view, update)

### Moderator Role

Có permissions hạn chế:

- users.view, users.update
- profile.view, profile.update

### User Role

Chỉ có basic permissions:

- profile.view, profile.update

## Check Permissions trong Code

```go
func DeleteUser(w http.ResponseWriter, r *http.Request) {
    claims := jwt.GetClaimsFromContext(r.Context())

    // Get user permissions
    userID, _ := uuid.Parse(claims.UserID)
    userInfo, _ := authService.GetUserInfo(r.Context(), userID)

    // Check permission
    if !hasPermission(userInfo.Permissions, "users.delete") {
        response.Forbidden(w, lang, response.CodePermissionDenied)
        return
    }

    // Proceed...
}

func hasPermission(permissions []string, required string) bool {
    for _, p := range permissions {
        if p == required {
            return true
        }
    }
    return false
}
```

## Security Notes

1. **Password**: Được hash với bcrypt trước khi lưu DB
2. **Token**: Access token 15 phút, Refresh token 7 ngày
3. **Blacklist**: Logout adds token to Redis blacklist
4. **Middleware**: Auto verify JWT và blacklist check
5. **Role Check**: Middleware `RequireRole()` để restrict routes

## Error Codes

| Code                 | Status | Message (EN)        | Message (VI)         |
| -------------------- | ------ | ------------------- | -------------------- |
| INVALID_CREDENTIALS  | 401    | Invalid credentials | Sai email/password   |
| TOKEN_MISSING        | 401    | Token is required   | Thiếu token          |
| TOKEN_INVALID        | 401    | Invalid token       | Token không hợp lệ   |
| TOKEN_EXPIRED        | 401    | Token expired       | Token hết hạn        |
| ACCOUNT_DISABLED     | 403    | Account disabled    | Tài khoản bị vô hiệu |
| PERMISSION_DENIED    | 403    | Permission denied   | Không có quyền       |
| EMAIL_ALREADY_EXISTS | 409    | Email exists        | Email đã tồn tại     |

## See Also

- [JWT Package](../../pkg/jwt/README.md)
- [Response Package](../../pkg/response/README.md)
- [JWT Guide](../../docs/jwt-guide.md)
