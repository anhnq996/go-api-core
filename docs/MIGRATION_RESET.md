# Reset & Run Migrations

Hướng dẫn reset database và chạy lại migrations sau khi sắp xếp lại.

## Migrations Mới

Thứ tự migrations đã được sắp xếp lại:

1. **000001_create_roles_table.sql** - Bảng roles
2. **000002_create_permissions_table.sql** - Bảng permissions
3. **000003_create_role_has_permissions_table.sql** - Bảng quan hệ
4. **000004_create_users_table.sql** - Bảng users (đầy đủ tất cả fields)

## Thay đổi

✅ Bỏ tất cả triggers (GORM tự động handle updated_at)
✅ Gộp users table vào 1 file duy nhất
✅ Sắp xếp thứ tự dependencies đúng

## Reset Database

### Option 1: Rollback All Migrations

```bash
# Rollback tất cả migrations về version 0
go run cmd/migrate/main.go down

# Check version (should be 0)
go run cmd/migrate/main.go version
```

### Option 2: Drop & Recreate Database (Recommended)

```bash
# Connect to PostgreSQL
docker exec -it apicore-postgres psql -U postgres

# In psql:
DROP DATABASE IF EXISTS apicore;
CREATE DATABASE apicore;
\q
```

## Run Migrations

```bash
# Run all migrations
make migrate

# Hoặc
go run cmd/migrate/main.go up

# Check version (should be 4)
go run cmd/migrate/main.go version
```

## Run Seeders

```bash
# Run all seeders
make seed

# Hoặc
go run cmd/migrate/main.go seed
```

## Verify

```bash
# Check tables created
docker exec -it apicore-postgres psql -U postgres -d apicore -c "\dt"

# Should see:
# - roles
# - permissions
# - role_has_permissions
# - users
# - schema_migrations

# Check seeded data
docker exec -it apicore-postgres psql -U postgres -d apicore -c "SELECT name FROM roles;"

# Should see:
# - admin
# - moderator
# - user
```

## Quick Reset Script

```bash
# Complete reset & setup
docker exec -it apicore-postgres psql -U postgres -c "DROP DATABASE IF EXISTS apicore;"
docker exec -it apicore-postgres psql -U postgres -c "CREATE DATABASE apicore;"
make migrate
make seed
make watch
```

## Troubleshooting

### Problem: Migration version mismatch

```sql
-- Connect to database
\c apicore

-- Check current version
SELECT * FROM schema_migrations;

-- Force set version to 0 (careful!)
DELETE FROM schema_migrations;
```

### Problem: Foreign key errors

Đảm bảo chạy migrations theo đúng thứ tự:

1. roles (trước)
2. permissions
3. role_has_permissions
4. users (sau, vì có FK to roles)

### Problem: Duplicate key errors

```bash
# Drop all tables and run again
go run cmd/migrate/main.go down
go run cmd/migrate/main.go up
```

## Final Check

Sau khi reset & migrate, test login:

```bash
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "Password123!"
  }'
```

Nếu trả về tokens → Setup thành công! ✅
