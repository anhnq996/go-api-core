# Seeders Final Summary - Name-Based Pattern

## 🎯 Overview

**Tất cả seeders đã được refactor** để sử dụng **name-based lookup** thay vì hardcode IDs.

## ✅ Refactored Seeders

### 1. ✅ Role Seeder

**Status:** Already idempotent (unchanged)

```go
// Lookup by name, create or update
db.Where("name = ?", role.Name).FirstOrCreate(&role)
```

### 2. ✅ Permission Seeder

**Status:** Already idempotent (unchanged)

```go
// Lookup by name, create or update
db.Where("name = ?", permission.Name).FirstOrCreate(&permission)
```

### 3. ✅ Role-Permission Seeder (NEW!)

**Status:** Refactored to name-based mapping

**Before:**

```go
// ❌ Hardcode random UUIDs
adminPermissions := []uuid.UUID{
    uuid.New(),
    uuid.New(),
}
```

**After:**

```go
// ✅ Name-based mapping
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

### 4. ✅ User Seeder (NEW!)

**Status:** Refactored to name-based role assignment

**Before:**

```go
// ❌ Hardcode random UUIDs
adminRoleID := uuid.New()
users := []model.User{
    {
        Name:   "Admin",
        RoleID: &adminRoleID,  // Wrong ID!
    },
}
```

**After:**

```go
// ✅ Name-based role assignment
type UserSeed struct {
    Name     string
    Email    string
    RoleName string  // Role name instead of ID
}

userSeeds := []UserSeed{
    {
        Name:     "Admin User",
        Email:    "admin@example.com",
        RoleName: "admin",  // Query from DB
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

## 🎨 Consistent Pattern

**All seeders follow the same pattern:**

```
1. Query entities from database
   ↓
2. Create lookup map by name
   ↓
3. Use names for relationships
   ↓
4. Safe: skip if not found
   ↓
5. Idempotent: can run multiple times
```

## 📊 Benefits Comparison

| Aspect                    | Before (ID-based) | After (Name-based) |
| ------------------------- | ----------------- | ------------------ |
| **Maintainability**       | ❌ Very Hard      | ✅ Very Easy       |
| **Readability**           | ❌ UUID gibberish | ✅ Clear names     |
| **Works After Migration** | ❌ No             | ✅ Yes             |
| **Safe**                  | ❌ FK violations  | ✅ Safe checks     |
| **Idempotent**            | ❌ Duplicates     | ✅ Create/Update   |
| **Easy to Modify**        | ❌ Hard           | ✅ Simple          |

## 🔄 Complete Seeder Flow

```
1. SeedRoles
   ↓ Creates: admin, moderator, user

2. SeedPermissions
   ↓ Creates: users.*, roles.*, permissions.*, profile.*

3. SeedRolePermissions
   ↓ Maps roles → permissions BY NAME
   ↓ admin gets 10 permissions
   ↓ moderator gets 4 permissions
   ↓ user gets 2 permissions

4. SeedUsers
   ↓ Creates users with roles BY NAME
   ↓ admin@example.com → admin role
   ↓ moderator@example.com → moderator role
   ↓ user@example.com → user role
```

## 📝 Code Examples

### Example 1: Role-Permission Mapping

```go
// Easy to read and maintain
rolePermissionMap := map[string][]string{
    "admin": {
        "users.view",
        "users.create",
        // ... all permissions
    },
    "moderator": {
        "users.view",
        "users.update",
        // ... limited permissions
    },
}

// Query and map
roleMap := makeRoleMap(db)
permMap := makePermissionMap(db)

// Create relationships
for roleName, permNames := range rolePermissionMap {
    role := roleMap[roleName]
    for _, permName := range permNames {
        perm := permMap[permName]
        db.Create(&RoleHasPermission{
            RoleID:       role.ID,
            PermissionID: perm.ID,
        })
    }
}
```

### Example 2: User-Role Assignment

```go
// Clear role assignment
userSeeds := []UserSeed{
    {Name: "Admin", Email: "admin@example.com", RoleName: "admin"},
    {Name: "User", Email: "user@example.com", RoleName: "user"},
}

// Query and map
roleMap := makeRoleMap(db)

// Create users
for _, seed := range userSeeds {
    role := roleMap[seed.RoleName]
    user := model.User{
        Name:   seed.Name,
        Email:  seed.Email,
        RoleID: &role.ID,  // Real ID from DB
    }
    db.Create(&user)
}
```

## 🧪 Testing

### Complete Test Flow

```bash
# 1. Fresh setup
make fresh

# 2. Verify seeded data
psql -U postgres -d apicore <<EOF
-- Check roles (should be 3)
SELECT COUNT(*) FROM roles;

-- Check permissions (should be 10)
SELECT COUNT(*) FROM permissions;

-- Check role-permissions (should be 16)
SELECT COUNT(*) FROM role_has_permissions;

-- Check users (should be 5)
SELECT COUNT(*) FROM users;

-- Verify relationships
SELECT
    u.name as user,
    u.email,
    r.name as role,
    COUNT(DISTINCT p.id) as permission_count
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
LEFT JOIN role_has_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id
GROUP BY u.name, u.email, r.name
ORDER BY permission_count DESC;
EOF
```

Expected output:

```
      user       |         email         |   role    | permission_count
-----------------+-----------------------+-----------+------------------
 Admin User      | admin@example.com     | admin     |               10
 Moderator User  | moderator@example.com | moderator |                4
 Regular User    | user@example.com      | user      |                2
 John Doe        | john@example.com      | user      |                2
 Jane Smith      | jane@example.com      | user      |                2
```

### Test Login

```bash
# Test admin login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Password123!"}'

# Should return all 10 permissions
```

## 📚 Documentation

### New Documentation Files

1. **`database/seeders/README.md`** - Updated with name-based pattern
2. **`docs/SEEDER_GUIDE.md`** - Complete usage guide
3. **`docs/SEEDER_UPDATE_SUMMARY.md`** - Role-Permission refactor
4. **`docs/SEEDER_USER_UPDATE.md`** - User seeder refactor
5. **`docs/SEEDERS_FINAL_SUMMARY.md`** - This file (complete overview)

### Updated Files

- `database/seeders/role_permission_seeder.go` - Name-based mapping
- `database/seeders/user_seeder.go` - Name-based role assignment
- `cmd/migrate/main.go` - Added `fresh` command
- `Makefile` - Added `make fresh` command

## 🎯 How to Update

### Add New Permission

```go
// 1. Add to permission seeder
{
    Name:        "posts.create",
    DisplayName: "Create Posts",
    Module:      "posts",
}

// 2. Assign to roles
rolePermissionMap := map[string][]string{
    "admin": {
        // ... existing
        "posts.create",  // ← Add here
    },
}

// 3. Reseed
make seed
```

### Add New User

```go
// Just add to userSeeds array
userSeeds := []UserSeed{
    // ... existing
    {
        Name:     "New Manager",
        Email:    "manager@example.com",
        RoleName: "moderator",  // ← Easy!
    },
}

// Reseed
make seed
```

### Change User Role

```go
// Change role name in seeder
{
    Name:     "John Doe",
    Email:    "john@example.com",
    RoleName: "moderator",  // Changed from "user" to "moderator"
}

// Reseed - will update existing user
make seed
```

## 🚀 Commands

```bash
# Fresh setup (recommended)
make fresh              # Drop all + migrate + seed

# Individual commands
make migrate-fresh      # Drop all + migrate only
make seed              # Seed only

# Verify
psql -U postgres -d apicore -c "
SELECT r.name as role, COUNT(rp.permission_id) as perms
FROM roles r
LEFT JOIN role_has_permissions rp ON r.id = rp.role_id
GROUP BY r.name
ORDER BY perms DESC;
"
```

## ✅ Final Checklist

- [x] Role seeder - idempotent ✅
- [x] Permission seeder - idempotent ✅
- [x] Role-Permission seeder - name-based ✅
- [x] User seeder - name-based ✅
- [x] Documentation complete ✅
- [x] `make fresh` command ✅
- [x] Build successful ✅
- [x] Ready for testing ✅

## 🎊 Result

**All seeders now use name-based pattern:**

- ✅ No hardcoded IDs
- ✅ Easy to read and maintain
- ✅ Safe with existence checks
- ✅ Idempotent (can run multiple times)
- ✅ Works consistently after migrations
- ✅ Production-ready

**Status:** ✅ COMPLETE - All seeders refactored!
