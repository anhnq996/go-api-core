# Setup Authentication Module

Hướng dẫn setup module authentication từ đầu.

## Bước 1: Run Migrations

```bash
# Run tất cả migrations
make migrate

# Hoặc
go run cmd/migrate/main.go up
```

Migrations sẽ tạo các bảng:

- ✅ roles
- ✅ permissions
- ✅ role_has_permissions
- ✅ users (updated với password, avatar, role_id)

## Bước 2: Run Seeders

```bash
# Run tất cả seeders
make seed

# Hoặc
go run cmd/migrate/main.go seed
```

Seeders sẽ tạo:

- ✅ 3 roles (admin, moderator, user)
- ✅ 10 permissions
- ✅ Role-permission relationships
- ✅ 5 demo users với passwords

## Bước 3: Configure JWT

Cập nhật file `.env`:

```env
JWT_SECRET_KEY=your-super-secret-key-at-least-32-characters-long-change-this
JWT_ACCESS_TOKEN_DURATION=15m
JWT_REFRESH_TOKEN_DURATION=168h
```

⚠️ **Quan trọng**: Đổi `JWT_SECRET_KEY` thành key mạnh và unique!

### Generate Strong Secret Key

```bash
# Method 1: OpenSSL
openssl rand -base64 48

# Method 2: Go
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 48); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'
```

## Bước 4: Start Server

```bash
# Start với hot reload
make watch

# Hoặc run bình thường
make run
```

## Bước 5: Test Login

```bash
# Login as admin
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "Password123!"
  }'
```

Response sẽ trả về:

- access_token
- refresh_token
- user info với permissions

## Test Accounts

| Email                 | Password     | Role      | Permissions                          |
| --------------------- | ------------ | --------- | ------------------------------------ |
| admin@example.com     | Password123! | admin     | ALL                                  |
| moderator@example.com | Password123! | moderator | users.view, users.update, profile.\* |
| user@example.com      | Password123! | user      | profile.\* only                      |

## API Endpoints

### Public (không cần auth)

- POST `/api/v1/auth/login` - Login
- POST `/api/v1/auth/register` - Register
- POST `/api/v1/auth/refresh` - Refresh token

### Protected (cần JWT token)

- GET `/api/v1/auth/me` - Get current user
- POST `/api/v1/auth/logout` - Logout current device
- POST `/api/v1/auth/logout-all` - Logout all devices
- ALL `/api/v1/users/*` - User management (protected)

## Workflow

```
1. Client login → Server verify credentials
2. Server generate JWT (access + refresh tokens)
3. Server return tokens + user info + permissions
4. Client save tokens (memory/cookie)
5. Client send token in Authorization header
6. Middleware verify token & check blacklist
7. Middleware add user info to context
8. Handler access user info from context
```

## Troubleshooting

### Problem: Migration failed

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Check connection
psql -h localhost -U postgres -d apicore

# Reset database (careful!)
make migrate-down
make migrate
```

### Problem: Seeder failed

```bash
# Run seeders individually
go run cmd/migrate/main.go up  # First
# Then manually run seeders one by one in order:
# 1. Roles
# 2. Permissions
# 3. Role-Permissions
# 4. Users
```

### Problem: Login failed

```bash
# Check user exists
psql -d apicore -c "SELECT email, is_active FROM users WHERE email='admin@example.com';"

# Check password is hashed
psql -d apicore -c "SELECT email, password FROM users LIMIT 1;"

# Check role assigned
psql -d apicore -c "SELECT u.email, r.name as role FROM users u LEFT JOIN roles r ON u.role_id = r.id;"
```

## Next Steps

1. ✅ Test all auth endpoints
2. ✅ Test with different roles
3. ✅ Implement permission checks in controllers
4. ✅ Add email verification (optional)
5. ✅ Add password reset (optional)
6. ✅ Add 2FA (optional)

## Files Created

### Migrations

- `database/migrations/000002_create_roles_table.up.sql`
- `database/migrations/000003_create_permissions_table.up.sql`
- `database/migrations/000004_create_role_has_permissions_table.up.sql`
- `database/migrations/000005_alter_users_table.up.sql`

### Models

- `internal/models/role.go`
- `internal/models/permission.go`
- `internal/models/role_permission.go`
- `internal/models/user.go` (updated)

### Repositories

- `internal/repositories/auth_repository.go`

### Auth Module

- `internal/app/auth/service.go`
- `internal/app/auth/controller.go`
- `internal/app/auth/route.go`

### Seeders

- `database/seeders/role_seeder.go`
- `database/seeders/permission_seeder.go`
- `database/seeders/role_permission_seeder.go`
- `database/seeders/user_seeder.go` (updated)

### Wire

- `internal/wire/providers.go` (new)
- `internal/wire/wire.go` (updated)
- `internal/wire/wire_gen.go` (updated)

### Routes

- `internal/routes/routes.go` (updated)
