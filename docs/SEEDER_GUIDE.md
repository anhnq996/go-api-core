# Seeder Guide

Hướng dẫn sử dụng seeders để khởi tạo dữ liệu mẫu.

## Quick Start

### 1. Reset Database & Seed

```bash
# Xóa tất cả tables và tạo lại
make migrate-fresh

# Chạy seeders
make seed
```

### 2. Verify Data

```bash
# Login với admin account
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "Password123!"
  }'
```

## Seeder Architecture (Updated!)

### Role-Permission Mapping

**Seeder sử dụng tên thay vì ID:**

```go
// database/seeders/role_permission_seeder.go
rolePermissionMap := map[string][]string{
    "admin": {
        "users.view",
        "users.create",
        "users.update",
        "users.delete",
        "roles.view",
        "roles.manage",
        "permissions.view",
        "permissions.manage",
        "profile.view",
        "profile.update",
    },
    "moderator": {
        "users.view",
        "users.update",
        "profile.view",
        "profile.update",
    },
    "user": {
        "profile.view",
        "profile.update",
    },
}
```

### Cách Hoạt Động

1. **Query theo tên từ database:**

```go
var roles []model.Role
db.Find(&roles) // Get all roles

var permissions []model.Permission
db.Find(&permissions) // Get all permissions
```

2. **Tạo lookup maps:**

```go
roleMap := make(map[string]model.Role)
for _, role := range roles {
    roleMap[role.Name] = role // Map by name
}

permissionMap := make(map[string]model.Permission)
for _, permission := range permissions {
    permissionMap[permission.Name] = permission // Map by name
}
```

3. **Tạo relationships:**

```go
for roleName, permissionNames := range rolePermissionMap {
    role := roleMap[roleName]

    for _, permName := range permissionNames {
        permission := permissionMap[permName]

        // Create relationship using actual IDs from database
        db.Create(&RoleHasPermission{
            RoleID:       role.ID,        // Real ID from DB
            PermissionID: permission.ID,   // Real ID from DB
        })
    }
}
```

## Test Accounts

### Admin Account

```json
{
  "email": "admin@example.com",
  "password": "Password123!",
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
}
```

### Moderator Account

```json
{
  "email": "moderator@example.com",
  "password": "Password123!",
  "permissions": [
    "users.view",
    "users.update",
    "profile.view",
    "profile.update"
  ]
}
```

### User Account

```json
{
  "email": "user@example.com",
  "password": "Password123!",
  "permissions": ["profile.view", "profile.update"]
}
```

## Testing Permissions

### Test Admin (Full Access)

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Password123!"}' \
  | jq -r '.data.access_token')

# Get profile (should have all permissions)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/api/v1/auth/me | jq '.data.permissions'

# Expected output:
# [
#   "users.view", "users.create", "users.update", "users.delete",
#   "roles.view", "roles.manage",
#   "permissions.view", "permissions.manage",
#   "profile.view", "profile.update"
# ]
```

### Test Moderator (Limited Access)

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@example.com","password":"Password123!"}' \
  | jq -r '.data.access_token')

# Get profile
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/api/v1/auth/me | jq '.data.permissions'

# Expected output:
# [
#   "users.view", "users.update",
#   "profile.view", "profile.update"
# ]
```

### Test User (Basic Access)

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password123!"}' \
  | jq -r '.data.access_token')

# Get profile
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/api/v1/auth/me | jq '.data.permissions'

# Expected output:
# [
#   "profile.view", "profile.update"
# ]
```

## Updating Permissions

### Add Permission to Role

**Ví dụ:** Thêm `users.create` cho moderator

```go
// database/seeders/role_permission_seeder.go
rolePermissionMap := map[string][]string{
    "moderator": {
        "users.view",
        "users.update",
        "users.create",     // ← Thêm permission mới
        "profile.view",
        "profile.update",
    },
}
```

**Áp dụng:**

```bash
# Chạy lại seeder
make seed

# Hoặc chỉ chạy role-permission seeder
go run cmd/migrate/main.go # (modify to run specific seeder)
```

### Remove Permission from Role

**Ví dụ:** Xóa `users.update` khỏi moderator

```go
rolePermissionMap := map[string][]string{
    "moderator": {
        "users.view",
        // "users.update",  ← Comment hoặc xóa
        "profile.view",
        "profile.update",
    },
}
```

### Add New Permission

**1. Thêm vào permission seeder:**

```go
// database/seeders/permission_seeder.go
permissions := []model.Permission{
    // ... existing permissions
    {
        Name:        "posts.create",
        DisplayName: "Create Posts",
        Description: "Can create new posts",
        Module:      "posts",
    },
}
```

**2. Assign to roles:**

```go
// database/seeders/role_permission_seeder.go
rolePermissionMap := map[string][]string{
    "admin": {
        // ... existing permissions
        "posts.create", // ← Add new permission
    },
}
```

**3. Run seeders:**

```bash
make seed
```

## SQL Queries for Debugging

### Check All Roles

```sql
SELECT id, name, display_name FROM roles ORDER BY name;
```

### Check All Permissions

```sql
SELECT id, name, display_name, module FROM permissions ORDER BY module, name;
```

### Check Role-Permission Relationships

```sql
SELECT
    r.name as role,
    p.name as permission,
    p.module
FROM role_has_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.module, p.name;
```

Expected output:

```
     role     |    permission     |   module
--------------+-------------------+-------------
 admin        | permissions.manage| permissions
 admin        | permissions.view  | permissions
 admin        | profile.update    | profile
 admin        | profile.view      | profile
 admin        | roles.manage      | roles
 admin        | roles.view        | roles
 admin        | users.create      | users
 admin        | users.delete      | users
 admin        | users.update      | users
 admin        | users.view        | users
 moderator    | profile.update    | profile
 moderator    | profile.view      | profile
 moderator    | users.update      | users
 moderator    | users.view        | users
 user         | profile.update    | profile
 user         | profile.view      | profile
```

### Count Permissions per Role

```sql
SELECT
    r.name as role,
    COUNT(rp.permission_id) as permission_count
FROM roles r
LEFT JOIN role_has_permissions rp ON r.id = rp.role_id
GROUP BY r.name
ORDER BY permission_count DESC;
```

Expected:

```
    role     | permission_count
-------------+-----------------
 admin       |               10
 moderator   |                4
 user        |                2
```

### Check User Roles

```sql
SELECT
    u.name,
    u.email,
    r.name as role,
    r.display_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
ORDER BY r.name;
```

## Common Issues

### Issue: No permissions returned after login

**Check:**

```sql
-- 1. User có role không?
SELECT u.email, r.name as role
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
WHERE u.email = 'admin@example.com';

-- 2. Role có permissions không?
SELECT p.name
FROM role_has_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'admin';
```

**Fix:**

```bash
# Reseed
make seed
```

### Issue: Duplicate key error khi seed

**Cause:** Role-permission relationships đã tồn tại.

**Fix:** Seeder tự động clear trước khi seed:

```go
// database/seeders/role_permission_seeder.go
db.Where("1 = 1").Delete(&model.RoleHasPermission{})
```

Nếu vẫn lỗi:

```bash
# Manual clear
psql -U postgres -d apicore -c "TRUNCATE role_has_permissions;"

# Reseed
make seed
```

### Issue: Permission không được assign

**Check tên có đúng không:**

```go
// ❌ Wrong - typo
"user.view" // missing 's'

// ✅ Correct
"users.view"
```

**Verify trong database:**

```sql
SELECT name FROM permissions WHERE name LIKE 'users%';
```

## Best Practices

### 1. Luôn dùng tên, không dùng ID

```go
// ✅ Good - Maintainable
rolePermissionMap := map[string][]string{
    "admin": {"users.view", "users.create"},
}

// ❌ Bad - Brittle, IDs change
adminPermissions := []uuid.UUID{
    uuid.MustParse("123e4567-..."), // This will break!
}
```

### 2. Group permissions theo module

```go
// ✅ Good - Organized
permissions := []model.Permission{
    // Users module
    {Name: "users.view", Module: "users"},
    {Name: "users.create", Module: "users"},

    // Posts module
    {Name: "posts.view", Module: "posts"},
    {Name: "posts.create", Module: "posts"},
}
```

### 3. Naming convention

```
{resource}.{action}

Examples:
- users.view
- users.create
- posts.publish
- reports.export
```

### 4. Role hierarchy

```
admin > moderator > user

Admin: All permissions
Moderator: Subset of admin
User: Basic permissions only
```

## Advanced Usage

### Custom Seeder Script

```bash
#!/bin/bash
# scripts/seed-test-data.sh

echo "🌱 Seeding test data..."

# Reset database
make migrate-fresh

# Seed base data
make seed

# Add custom test data
psql -U postgres -d apicore <<EOF
-- Insert additional test users
INSERT INTO users (name, email, password, role_id)
SELECT
    'Test User ' || i,
    'test' || i || '@example.com',
    '\$2a\$10\$hashed_password',
    (SELECT id FROM roles WHERE name = 'user')
FROM generate_series(1, 10) i;
EOF

echo "✅ Seeding complete!"
```

### Seeder cho Testing

```go
// database/seeders/test_seeder.go
func SeedTestData(db *gorm.DB) error {
    // Only for testing environment
    if os.Getenv("APP_ENV") != "testing" {
        return nil
    }

    // Seed minimal data for tests
    // ...
}
```

## Resources

- [Database Seeders README](../database/seeders/README.md)
- [Migration Guide](../database/migrations/README.md)
- [Auth Module](AUTH_README.md)
