# Database Seeders

Seeders để khởi tạo dữ liệu mẫu cho database.

## Seeder Order

Seeders chạy theo thứ tự trong `seeder.go`:

```go
var AllSeeders = []Seeder{
    {"Roles", SeedRoles},
    {"Permissions", SeedPermissions},
    {"RolePermissions", SeedRolePermissions},
    {"Users", SeedUsers},
}
```

**⚠️ Quan trọng:** Thứ tự này phải được tuân thủ vì có dependencies:

1. `Roles` - Tạo roles trước
2. `Permissions` - Tạo permissions
3. `RolePermissions` - Map roles với permissions (cần roles và permissions đã tồn tại)
4. `Users` - Tạo users (cần roles đã tồn tại)

## Seeders

### 1. Role Seeder

**File:** `role_seeder.go`

Tạo 3 roles cơ bản:

| Name        | Display Name  | Description                             |
| ----------- | ------------- | --------------------------------------- |
| `admin`     | Administrator | Full system access with all permissions |
| `moderator` | Moderator     | Can manage content and users            |
| `user`      | User          | Regular user with basic permissions     |

**Idempotent:** Nếu role đã tồn tại (theo `name`), sẽ update thay vì tạo mới.

### 2. Permission Seeder

**File:** `permission_seeder.go`

Tạo các permissions theo module:

#### Users Module

- `users.view` - View user list and details
- `users.create` - Create new users
- `users.update` - Update user information
- `users.delete` - Delete users

#### Roles Module

- `roles.view` - View roles
- `roles.manage` - Create, update, delete roles

#### Permissions Module

- `permissions.view` - View permissions
- `permissions.manage` - Assign/revoke permissions

#### Profile Module

- `profile.view` - View own profile
- `profile.update` - Update own profile

**Idempotent:** Nếu permission đã tồn tại (theo `name`), sẽ update thay vì tạo mới.

### 3. Role-Permission Seeder (Updated!)

**File:** `role_permission_seeder.go`

**🎯 Sử dụng tên thay vì ID để map relationships:**

```go
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

**Cách hoạt động:**

1. Query tất cả roles và permissions từ database
2. Tạo map[name]model để lookup nhanh
3. Dựa vào `rolePermissionMap`, tạo relationships theo tên
4. Nếu role hoặc permission không tồn tại, skip (không báo lỗi)

**Ưu điểm:**

- ✅ Không phụ thuộc vào ID (UUID thay đổi mỗi lần migrate)
- ✅ Dễ đọc và maintain
- ✅ Dễ thêm/xóa permissions cho roles
- ✅ Safe: skip nếu role/permission không tồn tại

**Cập nhật permissions cho role:**

```go
// Thêm permission mới cho moderator
"moderator": {
    "users.view",
    "users.update",
    "users.create",     // ← Thêm permission mới
    "profile.view",
    "profile.update",
},
```

### 4. User Seeder (Updated!)

**File:** `user_seeder.go`

**🎯 Sử dụng role name thay vì ID:**

```go
type UserSeed struct {
    Name     string
    Email    string
    RoleName string  // ✅ Dùng tên role thay vì ID
}

userSeeds := []UserSeed{
    {
        Name:     "Admin User",
        Email:    "admin@example.com",
        RoleName: "admin",  // ✅ Query role từ DB
    },
    // ...
}
```

**Cách hoạt động:**

1. Query all roles từ database
2. Tạo roleMap[name] để lookup
3. Assign role ID dựa trên role name

Tạo 5 users mẫu với mật khẩu đã hash:

| Email                 | Name           | Password     | Role      |
| --------------------- | -------------- | ------------ | --------- |
| admin@example.com     | Admin User     | Password123! | admin     |
| moderator@example.com | Moderator User | Password123! | moderator |
| user@example.com      | Regular User   | Password123! | user      |
| john@example.com      | John Doe       | Password123! | user      |
| jane@example.com      | Jane Smith     | Password123! | user      |

**Idempotent:** Nếu user đã tồn tại (theo `email`), sẽ update thay vì tạo mới.

## Usage

### Run All Seeders

```bash
make seed
# hoặc
go run cmd/migrate/main.go seed
```

### Run Specific Seeder

Sửa `cmd/migrate/main.go` hoặc tạo custom command:

```go
// Chỉ seed roles và permissions
seeders := []seeders.Seeder{
    {"Roles", seeders.SeedRoles},
    {"Permissions", seeders.SeedPermissions},
}
```

### Reset & Reseed

```bash
# Drop all tables, migrate, seed
make migrate-fresh
make seed
```

## Adding New Seeders

### 1. Tạo seeder file mới

```go
// database/seeders/product_seeder.go
package seeders

import (
    model "anhnq/api-core/internal/models"
    "gorm.io/gorm"
)

func SeedProducts(db *gorm.DB) error {
    products := []model.Product{
        {
            Name:  "Product 1",
            Price: 100.00,
        },
        // ...
    }

    for _, product := range products {
        var existing model.Product
        if err := db.Where("name = ?", product.Name).First(&existing).Error; err != nil {
            // Create new
            if err := db.Create(&product).Error; err != nil {
                return err
            }
        } else {
            // Update existing
            product.ID = existing.ID
            if err := db.Model(&existing).Updates(product).Error; err != nil {
                return err
            }
        }
    }

    return nil
}
```

### 2. Register trong `seeder.go`

```go
var AllSeeders = []Seeder{
    {"Roles", SeedRoles},
    {"Permissions", SeedPermissions},
    {"RolePermissions", SeedRolePermissions},
    {"Users", SeedUsers},
    {"Products", SeedProducts}, // ← Thêm vào đây
}
```

## Best Practices

### 1. Idempotent Seeders

Luôn check exist trước khi create:

```go
// ✅ Good - Idempotent
var existing model.Role
if err := db.Where("name = ?", role.Name).First(&existing).Error; err != nil {
    // Create
    db.Create(&role)
} else {
    // Update
    db.Model(&existing).Updates(role)
}

// ❌ Bad - Sẽ lỗi nếu chạy lại
db.Create(&role) // Duplicate key error
```

### 2. Sử dụng Name/Slug thay vì ID

```go
// ✅ Good - Dùng tên
rolePermissionMap := map[string][]string{
    "admin": {"users.view", "users.create"},
}

// ❌ Bad - Hardcode ID
adminPermissions := []uuid.UUID{
    uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
}
```

### 3. Clear relationships trước khi seed

```go
// Clear existing để tránh duplicates
db.Where("1 = 1").Delete(&model.RoleHasPermission{})

// Sau đó seed mới
for _, mapping := range mappings {
    db.Create(&mapping)
}
```

### 4. Handle errors gracefully

```go
// Skip nếu không tìm thấy
role, exists := roleMap[roleName]
if !exists {
    continue // Skip instead of error
}
```

### 5. Order matters

Dependencies phải được seed trước:

```
Roles → Permissions → RolePermissions → Users
  ↑                                         ↓
  └─────────────────────────────────────────┘
  (Users cần RoleID)
```

## Testing

### Check seeded data

```bash
# Connect to database
psql -U postgres -d apicore

# Check roles
SELECT * FROM roles;

# Check permissions
SELECT * FROM permissions;

# Check role-permission relationships
SELECT r.name as role, p.name as permission
FROM role_has_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.name;

# Check users
SELECT name, email, r.name as role
FROM users u
LEFT JOIN roles r ON u.role_id = r.id;
```

### Verify admin permissions

```sql
SELECT p.name
FROM role_has_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'admin';
```

Should return 10 permissions (all).

## Troubleshooting

### Problem: "role_id" violates foreign key constraint

**Cause:** Trying to create user before roles exist.

**Fix:** Ensure seeder order is correct (Roles → Users).

### Problem: Duplicate key error

**Cause:** Seeder không idempotent, chạy lại tạo duplicate.

**Fix:** Check exist before create:

```go
if err := db.Where("name = ?", name).First(&existing).Error; err != nil {
    db.Create(&item)
}
```

### Problem: Role-permission relationships không tạo

**Cause:** Role hoặc Permission không tồn tại trong database.

**Fix:**

1. Check `SeedRoles` và `SeedPermissions` chạy thành công
2. Check tên trong `rolePermissionMap` match với database
3. Add logging:

```go
if !roleExists {
    log.Printf("Role '%s' not found, skipping", roleName)
    continue
}
```

## Migration vs Seeding

| Migration                    | Seeding                 |
| ---------------------------- | ----------------------- |
| Schema changes (DDL)         | Test/initial data (DML) |
| CREATE TABLE, ALTER TABLE    | INSERT, UPDATE          |
| Always run in production     | Optional in production  |
| Version controlled (up/down) | Can run multiple times  |
| Required for app to work     | Optional convenience    |

## Example: Complete Setup

```bash
# 1. Reset database
make migrate-fresh

# 2. Run seeders
make seed

# 3. Verify
psql -U postgres -d apicore -c "SELECT COUNT(*) FROM roles;"
# Should return: 3

psql -U postgres -d apicore -c "SELECT COUNT(*) FROM permissions;"
# Should return: 10

psql -U postgres -d apicore -c "SELECT COUNT(*) FROM role_has_permissions;"
# Should return: 16 (10 admin + 4 moderator + 2 user)

psql -U postgres -d apicore -c "SELECT COUNT(*) FROM users;"
# Should return: 3

# 4. Test login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Password123!"}'
```

## See Also

- [Migration Guide](../migrations/README.md)
- [Models Documentation](../../internal/models/)
