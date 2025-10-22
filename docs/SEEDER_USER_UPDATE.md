# User Seeder Update

## 🎯 Problem

User seeder cũ hardcode UUID random cho role IDs:

```go
// ❌ OLD - WRONG
adminRoleID := uuid.New()     // Random UUID, không tồn tại trong DB!
moderatorRoleID := uuid.New()
userRoleID := uuid.New()

users := []model.User{
    {
        Name:     "Admin User",
        Email:    "admin@example.com",
        RoleID:   &adminRoleID,  // ❌ Sai! ID này không có trong DB
    },
}
```

**Vấn đề:**

- ❌ Role IDs không match với database
- ❌ Users được tạo nhưng không có role thực tế
- ❌ Foreign key constraint có thể fail
- ❌ Không maintain được

## ✅ Solution

Sử dụng **name-based role lookup**:

```go
// ✅ NEW - CORRECT
type UserSeed struct {
    Name     string
    Email    string
    RoleName string  // Tên role thay vì ID
}

userSeeds := []UserSeed{
    {
        Name:     "Admin User",
        Email:    "admin@example.com",
        RoleName: "admin",  // ✅ Query role từ DB theo tên
    },
    {
        Name:     "Moderator User",
        Email:    "moderator@example.com",
        RoleName: "moderator",
    },
    {
        Name:     "Regular User",
        Email:    "user@example.com",
        RoleName: "user",
    },
}
```

## 🔄 How It Works

### Step 1: Query roles từ database

```go
var roles []model.Role
db.Find(&roles) // Get all roles
```

### Step 2: Tạo lookup map

```go
roleMap := make(map[string]model.Role)
for _, role := range roles {
    roleMap[role.Name] = role
}
```

### Step 3: Assign role by name

```go
for _, userSeed := range userSeeds {
    // Lookup role by name
    role, roleExists := roleMap[userSeed.RoleName]
    if !roleExists {
        fmt.Printf("Role '%s' not found, skipping user\n", userSeed.RoleName)
        continue
    }

    // Create user với real role ID từ DB
    newUser := model.User{
        Name:     userSeed.Name,
        Email:    userSeed.Email,
        RoleID:   &role.ID,  // ✅ Real ID from database
    }

    db.Create(&newUser)
}
```

## 📊 Comparison

| Aspect                    | Old (Hardcode ID)      | New (Name-based)    |
| ------------------------- | ---------------------- | ------------------- |
| **Maintainability**       | ❌ Hard                | ✅ Easy             |
| **Readability**           | ❌ UUID không đọc được | ✅ Rõ ràng          |
| **Works After Migration** | ❌ No                  | ✅ Yes              |
| **Safe**                  | ❌ FK violation risk   | ✅ Safe with checks |

## 🎨 Benefits

### 1. Dễ đọc

```go
// Nhìn là biết user này có role gì
{
    Name:     "Admin User",
    Email:    "admin@example.com",
    RoleName: "admin",  // ✅ Clear!
}
```

### 2. Dễ maintain

```go
// Thêm user mới - chỉ cần specify role name
{
    Name:     "New Manager",
    Email:    "manager@example.com",
    RoleName: "moderator",  // ✅ Simple!
}
```

### 3. Safe

```go
// Kiểm tra role tồn tại
role, roleExists := roleMap[userSeed.RoleName]
if !roleExists {
    continue // Skip, không crash
}
```

### 4. Idempotent

```go
// Check exist trước khi create
var existingUser model.User
if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
    // Update existing
    db.Model(&existingUser).Updates(user)
} else {
    // Create new
    db.Create(&user)
}
```

## 📝 Updated Code

### Before

```go
// ❌ Hardcode UUID
adminRoleID := uuid.New()
users := []model.User{
    {
        Name:   "Admin",
        RoleID: &adminRoleID,
    },
}
```

### After

```go
// ✅ Name-based lookup
userSeeds := []UserSeed{
    {
        Name:     "Admin",
        RoleName: "admin",
    },
}

// Query and map
roleMap := make(map[string]model.Role)
for _, role := range roles {
    roleMap[role.Name] = role
}

// Assign real IDs
for _, seed := range userSeeds {
    role := roleMap[seed.RoleName]
    user := model.User{
        Name:   seed.Name,
        RoleID: &role.ID,  // Real ID
    }
}
```

## 🧪 Testing

### Test Seeder

```bash
# 1. Fresh setup
make fresh

# 2. Check users có role đúng không
psql -U postgres -d apicore -c "
SELECT
    u.name,
    u.email,
    r.name as role,
    r.display_name
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
ORDER BY r.name;
"
```

Expected output:

```
     name      |         email         |   role    | display_name
---------------+-----------------------+-----------+---------------
 Admin User    | admin@example.com     | admin     | Administrator
 Moderator User| moderator@example.com | moderator | Moderator
 Regular User  | user@example.com      | user      | User
 John Doe      | john@example.com      | user      | User
 Jane Smith    | jane@example.com      | user      | User
```

### Test Login

```bash
# Login với admin
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Password123!"}'

# Response should include role and permissions
{
  "success": true,
  "data": {
    "user": {
      "name": "Admin User",
      "email": "admin@example.com",
      "role": {
        "name": "admin",
        "display_name": "Administrator"
      }
    },
    "permissions": [...]
  }
}
```

## 🚀 Adding New Users

### Method 1: Trong Seeder

```go
userSeeds := []UserSeed{
    // Existing users...
    {
        Name:     "Support Agent",
        Email:    "support@example.com",
        RoleName: "moderator",  // ✅ Easy to add!
    },
}
```

### Method 2: API (Production)

```bash
# Use admin account to create users
curl -X POST http://localhost:3000/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "New User",
    "email": "newuser@example.com",
    "role_id": "role-uuid-from-db"
  }'
```

## 🎯 Summary

**Cả 2 seeders đã được refactor:**

1. **RolePermission Seeder**: Role → Permissions mapping by name ✅
2. **User Seeder**: User → Role assignment by name ✅

**Pattern nhất quán:**

```
Query from DB → Create lookup map → Use names → Safe & Maintainable
```

**Benefits:**

- ✅ Không phụ thuộc vào IDs
- ✅ Dễ đọc và maintain
- ✅ Safe với checks
- ✅ Idempotent
- ✅ Works after any migration

## 📚 Files Changed

- `database/seeders/user_seeder.go` - Refactored to name-based
- `database/seeders/README.md` - Updated documentation
- `docs/SEEDER_GUIDE.md` - Updated test accounts section
- `docs/SEEDER_USER_UPDATE.md` - This file (new)

## ✅ Checklist

- [x] Refactor user seeder to name-based
- [x] Update documentation
- [x] Test build
- [x] Ready to test with `make fresh`

**Status:** ✅ COMPLETE
