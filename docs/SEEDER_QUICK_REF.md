# Seeder Quick Reference

## 🚀 Quick Commands

```bash
# Fresh setup (recommended)
make fresh                    # Drop all → migrate → seed

# Individual operations
make migrate-fresh           # Drop all → migrate
make seed                    # Seed only

# Verify
psql -U postgres -d apicore -c "SELECT name, email, r.name as role FROM users u LEFT JOIN roles r ON u.role_id = r.id;"
```

## 📊 Seeded Data

### Roles (3)

- `admin` - Administrator (all permissions)
- `moderator` - Moderator (limited permissions)
- `user` - User (basic permissions)

### Permissions (10)

- **Users:** view, create, update, delete
- **Roles:** view, manage
- **Permissions:** view, manage
- **Profile:** view, update

### Users (5)

| Email                 | Password     | Role      | Permissions |
| --------------------- | ------------ | --------- | ----------- |
| admin@example.com     | Password123! | admin     | 10          |
| moderator@example.com | Password123! | moderator | 4           |
| user@example.com      | Password123! | user      | 2           |
| john@example.com      | Password123! | user      | 2           |
| jane@example.com      | Password123! | user      | 2           |

## 🎯 Pattern (Name-Based)

### Role-Permission Mapping

```go
rolePermissionMap := map[string][]string{
    "admin": {"users.view", "users.create", ...},
    "moderator": {"users.view", "users.update"},
    "user": {"profile.view", "profile.update"},
}
```

### User-Role Assignment

```go
userSeeds := []UserSeed{
    {Name: "Admin", Email: "admin@...", RoleName: "admin"},
    {Name: "User", Email: "user@...", RoleName: "user"},
}
```

## ✏️ How to Modify

### Add Permission to Role

```go
// database/seeders/role_permission_seeder.go
"moderator": {
    "users.view",
    "users.update",
    "users.create",  // ← Add this line
}
```

### Add New User

```go
// database/seeders/user_seeder.go
userSeeds := []UserSeed{
    // ... existing
    {
        Name:     "New User",
        Email:    "new@example.com",
        RoleName: "user",  // ← Add this
    },
}
```

### Add New Permission

```go
// 1. database/seeders/permission_seeder.go
{
    Name:        "posts.create",
    DisplayName: "Create Posts",
    Module:      "posts",
}

// 2. database/seeders/role_permission_seeder.go
"admin": {
    // ... existing
    "posts.create",  // ← Add to roles
}
```

**Then run:** `make seed`

## 🧪 Test Login

```bash
# Admin (all permissions)
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Password123!"}'

# Moderator (limited)
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"moderator@example.com","password":"Password123!"}'

# User (basic)
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Password123!"}'
```

## 🔍 SQL Queries

```sql
-- Check permission counts per role
SELECT
    r.name as role,
    COUNT(rp.permission_id) as permission_count
FROM roles r
LEFT JOIN role_has_permissions rp ON r.id = rp.role_id
GROUP BY r.name
ORDER BY permission_count DESC;

-- View all role-permission relationships
SELECT
    r.name as role,
    p.name as permission,
    p.module
FROM role_has_permissions rp
JOIN roles r ON rp.role_id = r.id
JOIN permissions p ON rp.permission_id = p.id
ORDER BY r.name, p.module, p.name;

-- Check users with roles
SELECT
    u.name,
    u.email,
    r.name as role
FROM users u
LEFT JOIN roles r ON u.role_id = r.id
ORDER BY r.name;
```

## 📚 Documentation

- [Complete Guide](SEEDER_GUIDE.md)
- [Pattern Details](SEEDERS_FINAL_SUMMARY.md)
- [Migration Guide](../database/migrations/README.md)
