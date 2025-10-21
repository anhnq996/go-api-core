# Migrations với golang-migrate

## Giới Thiệu

Sử dụng [golang-migrate](https://github.com/golang-migrate/migrate) để quản lý database migrations với version tracking.

## Tính Năng

- ✅ **Version tracking** - Track migration version trong DB
- ✅ **Up/Down migrations** - Migrate lên hoặc rollback
- ✅ **Steps** - Chạy N migrations
- ✅ **Force** - Fix dirty state
- ✅ **SQL files** - Migrations bằng pure SQL
- ✅ **Dirty detection** - Phát hiện migrations bị lỗi

## Cấu Trúc Files

```
database/migrations/
├── 000001_create_users_table.up.sql    # Migration up
├── 000001_create_users_table.down.sql  # Migration down
├── 000002_create_orders_table.up.sql   # Next migration
├── 000002_create_orders_table.down.sql
└── README.md
```

**Naming convention:**

- Format: `{version}_{description}.{up|down}.sql`
- Version: 6 digits (000001, 000002...)
- Description: snake_case
- Extension: `.up.sql` (migrate up), `.down.sql` (rollback)

## Commands

### Method 1: Using main.go flags

```bash
# Run all pending migrations
go run cmd/app/main.go -migrate

# Rollback all migrations
go run cmd/app/main.go -migrate-down

# Run 1 step up
go run cmd/app/main.go -migrate-steps=1

# Rollback 1 step
go run cmd/app/main.go -migrate-steps=-1
```

### Method 2: Using migrate CLI tool

```bash
# Run migrations up
go run cmd/migrate/main.go up

# Rollback all
go run cmd/migrate/main.go down

# Show version
go run cmd/migrate/main.go version

# Run 2 steps up
go run cmd/migrate/main.go steps -n 2

# Rollback 1 step
go run cmd/migrate/main.go steps -n -1

# Force version (when dirty)
go run cmd/migrate/main.go force -version 1
```

## Tạo Migration Mới

### Bước 1: Tạo files

**File:** `database/migrations/000002_create_orders_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    total DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    CONSTRAINT fk_orders_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at);
```

**File:** `database/migrations/000002_create_orders_table.down.sql`

```sql
DROP TABLE IF EXISTS orders;
```

### Bước 2: Run migration

```bash
go run cmd/migrate/main.go up
```

## Migration Versions

golang-migrate tự động track versions trong table `schema_migrations`:

```sql
-- Check migration version
SELECT * FROM schema_migrations;

-- Output:
-- version | dirty
-- --------|------
--    1    | false
```

## Dirty State

**Khi nào xảy ra?**

- Migration chạy bị lỗi giữa chừng
- Database trong trạng thái không consistent

**Cách fix:**

```bash
# 1. Check version
go run cmd/migrate/main.go version
# Output: version=1, dirty=true

# 2. Xem lỗi gì trong database

# 3. Fix manually hoặc force về version cũ
go run cmd/migrate/main.go force -version 0

# 4. Run lại migration
go run cmd/migrate/main.go up
```

## Best Practices

### 1. Luôn Test Migrations

```bash
# 1. Backup database
pg_dump apicore > backup.sql

# 2. Run migration
go run cmd/migrate/main.go up

# 3. Test app
go run cmd/app/main.go

# 4. Nếu có lỗi, rollback
go run cmd/migrate/main.go down

# 5. Restore backup nếu cần
psql apicore < backup.sql
```

### 2. Viết Idempotent Migrations

```sql
-- ✅ Good - idempotent
CREATE TABLE IF NOT EXISTS users (...);
CREATE INDEX IF NOT EXISTS idx_name ON users(name);
ALTER TABLE users ADD COLUMN IF NOT EXISTS age INT;

-- ❌ Bad - fail if run twice
CREATE TABLE users (...);
CREATE INDEX idx_name ON users(name);
```

### 3. Luôn Viết Down Migrations

Mỗi `.up.sql` phải có `.down.sql` tương ứng:

```sql
-- up: Add column
ALTER TABLE users ADD COLUMN age INT;

-- down: Remove column
ALTER TABLE users DROP COLUMN age;
```

### 4. Không Sửa Migrations Đã Chạy

```bash
# ❌ Không nên sửa file đã migrate
# ✅ Tạo migration mới để fix

# Example: Sửa column type
# Tạo: 000003_change_user_age_type.up.sql
ALTER TABLE users ALTER COLUMN age TYPE BIGINT;
```

## Examples

### Example 1: Add Column

**000003_add_user_phone.up.sql:**

```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
```

**000003_add_user_phone.down.sql:**

```sql
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users DROP COLUMN phone;
```

### Example 2: Add Table with Foreign Key

**000004_create_profiles_table.up.sql:**

```sql
CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    bio TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_profiles_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);
```

**000004_create_profiles_table.down.sql:**

```sql
DROP TABLE IF EXISTS profiles;
```

### Example 3: Add Index

**000005_add_user_name_index.up.sql:**

```sql
CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
```

**000005_add_user_name_index.down.sql:**

```sql
DROP INDEX IF EXISTS idx_users_name;
```

## Troubleshooting

### Migration failed: dirty state

```bash
# Check version and dirty state
go run cmd/migrate/main.go version
# Output: version=2, dirty=true

# Fix manually in database or force
go run cmd/migrate/main.go force -version 2

# Run again
go run cmd/migrate/main.go up
```

### No changes detected

```
Error: no change
```

This is normal - all migrations already run.

### Cannot find migrations directory

```bash
# Make sure you run from project root
cd /path/to/ApiCore
go run cmd/migrate/main.go up

# Or use absolute path in code
```

## Cheat Sheet

```bash
# Show current version
go run cmd/migrate/main.go version

# Run all migrations
go run cmd/migrate/main.go up

# Rollback all
go run cmd/migrate/main.go down

# Run 1 migration
go run cmd/migrate/main.go steps -n 1

# Rollback 1 migration
go run cmd/migrate/main.go steps -n -1

# Force version (emergency)
go run cmd/migrate/main.go force -version 1

# Check in database
psql apicore -c "SELECT * FROM schema_migrations;"
```

## Integration với GORM

golang-migrate tạo schema, GORM sử dụng schema đó:

```go
// Migration tạo table
-- 000001_create_users_table.up.sql

// GORM sử dụng
type User struct {
    ID string `gorm:"type:uuid;primaryKey"`
    // ... match với SQL schema
}

db.Find(&users) // GORM query từ table đã tạo
```

## Resources

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [Migration Best Practices](https://www.brunoscheufler.com/blog/2021-01-30-database-migration-strategies)
