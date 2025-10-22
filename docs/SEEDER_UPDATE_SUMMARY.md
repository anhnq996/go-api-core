# Seeder Update Summary

## 🎯 Problem

Seeder cũ (`role_permission_seeder.go`) hardcode UUID bằng `uuid.New()`, dẫn đến:

- ❌ IDs không match với database thực tế
- ❌ Relationships không được tạo đúng
- ❌ Mỗi lần chạy lại migration, IDs thay đổi
- ❌ Không thể maintain được

```go
// ❌ OLD - WRONG
adminPermissions := []uuid.UUID{
    uuid.New(),  // Random UUID, không tồn tại trong DB!
    uuid.New(),
    // ...
}
```

## ✅ Solution

Sử dụng **name-based mapping** thay vì ID:

```go
// ✅ NEW - CORRECT
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

## 🔄 How It Works

### Step 1: Query theo tên

```go
var roles []model.Role
db.Find(&roles) // Get all roles from DB

var permissions []model.Permission
db.Find(&permissions) // Get all permissions from DB
```

### Step 2: Tạo lookup maps

```go
roleMap := make(map[string]model.Role)
for _, role := range roles {
    roleMap[role.Name] = role
}

permissionMap := make(map[string]model.Permission)
for _, permission := range permissions {
    permissionMap[permission.Name] = permission
}
```

### Step 3: Tạo relationships

```go
for roleName, permissionNames := range rolePermissionMap {
    role := roleMap[roleName]

    for _, permName := range permissionNames {
        permission := permissionMap[permName]

        // Create relationship với IDs thực từ database
        db.Create(&RoleHasPermission{
            RoleID:       role.ID,        // ✅ Real ID
            PermissionID: permission.ID,   // ✅ Real ID
        })
    }
}
```

## 📊 Comparison

| Aspect                    | Old (UUID-based) | New (Name-based) |
| ------------------------- | ---------------- | ---------------- |
| **Maintainability**       | ❌ Hard          | ✅ Easy          |
| **Readability**           | ❌ Poor          | ✅ Excellent     |
| **ID Consistency**        | ❌ Random        | ✅ Consistent    |
| **Works After Migration** | ❌ No            | ✅ Yes           |
| **Easy to Update**        | ❌ No            | ✅ Yes           |

## 🎨 Benefits

### 1. Dễ đọc và hiểu

```go
// Nhìn là biết admin có quyền gì
"admin": {
    "users.view",
    "users.create",
    "users.update",
    "users.delete",
    // ...
}
```

### 2. Dễ maintain

```go
// Thêm permission cho moderator - chỉ cần thêm 1 dòng!
"moderator": {
    "users.view",
    "users.update",
    "users.create",     // ← Thêm permission mới
    "profile.view",
    "profile.update",
}
```

### 3. Safe

```go
// Nếu role hoặc permission không tồn tại, skip
role, exists := roleMap[roleName]
if !exists {
    continue // Skip, không crash
}
```

### 4. Idempotent

```go
// Clear trước khi seed, có thể chạy lại nhiều lần
db.Where("1 = 1").Delete(&model.RoleHasPermission{})
```

## 📝 Files Updated

### 1. `database/seeders/role_permission_seeder.go`

- ✅ Đổi từ hardcode UUID sang name-based mapping
- ✅ Query roles và permissions từ database
- ✅ Tạo lookup maps
- ✅ Safe handling khi không tìm thấy

### 2. `database/seeders/README.md`

- ✅ Document chi tiết về seeder pattern
- ✅ Hướng dẫn cách thêm/xóa permissions
- ✅ Best practices
- ✅ Troubleshooting guide

### 3. `docs/SEEDER_GUIDE.md`

- ✅ Quick start guide
- ✅ Testing guide với curl commands
- ✅ SQL queries for debugging
- ✅ Common issues và solutions

### 4. `cmd/migrate/main.go`

- ✅ Thêm `fresh` command
- ✅ Drop all + migrate + seed

### 5. `Makefile`

- ✅ Thêm `make fresh` command
- ✅ Update help text

## 🚀 Usage

### Quick Test

```bash
# 1. Fresh setup
make fresh

# 2. Login with admin
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Password123!"}'

# Response should include permissions:
{
  "success": true,
  "data": {
    "user": {
      "name": "Admin User",
      "email": "admin@example.com",
      "role": "admin"
    },
    "permissions": [
      "users.view", "users.create", "users.update", "users.delete",
      "roles.view", "roles.manage",
      "permissions.view", "permissions.manage",
      "profile.view", "profile.update"
    ]
  }
}
```

### Verify in Database

```sql
-- Check role-permission relationships
SELECT
    r.name as role,
    p.name as permission
FROM role_has_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.name;
```

Expected: 16 rows (10 admin + 4 moderator + 2 user)

## 📚 Documentation

- [Seeder README](../database/seeders/README.md) - Detailed seeder documentation
- [Seeder Guide](SEEDER_GUIDE.md) - Usage guide with examples
- [Migration Guide](../database/migrations/README.md) - Migration docs

## 🎯 Summary

**Before:**

```go
adminPermissions := []uuid.UUID{
    uuid.New(), // ❌ Wrong!
}
```

**After:**

```go
rolePermissionMap := map[string][]string{
    "admin": {"users.view", "users.create"}, // ✅ Correct!
}
```

**Result:**

- ✅ Seeder hoạt động đúng
- ✅ IDs được map từ database thực tế
- ✅ Dễ đọc, dễ maintain
- ✅ Safe và idempotent
- ✅ Works after fresh migration

## ✅ Checklist

- [x] Update `role_permission_seeder.go`
- [x] Add documentation
- [x] Add `make fresh` command
- [x] Update migration tool
- [x] Test với `make fresh`
- [x] Verify permissions sau login

**Status:** ✅ COMPLETE
