# Database Migrations

## Migration Files

Thứ tự chạy migrations:

1. **000001_create_roles_table.sql** - Tạo bảng roles
2. **000002_create_permissions_table.sql** - Tạo bảng permissions
3. **000003_create_role_has_permissions_table.sql** - Tạo bảng quan hệ role-permission
4. **000004_create_users_table.sql** - Tạo bảng users (với foreign key đến roles)

## Chạy Migrations

```bash
# Run all pending migrations
make migrate

# Hoặc
go run cmd/migrate/main.go up

# Rollback last migration
make migrate-down

# Check migration version
go run cmd/migrate/main.go version
```

## Reset Database

```bash
# Rollback tất cả migrations
go run cmd/migrate/main.go down

# Run lại tất cả
go run cmd/migrate/main.go up

# Run seeders
make seed
```

## Tạo Migration Mới

```bash
migrate create -ext sql -dir database/migrations -seq <migration_name>

# Ví dụ:
migrate create -ext sql -dir database/migrations -seq create_posts_table
```

## Schema

### roles

- id (UUID, PK)
- name (varchar(50), unique)
- display_name (varchar(100))
- description (text)
- created_at, updated_at

### permissions

- id (UUID, PK)
- name (varchar(100), unique)
- display_name (varchar(150))
- description (text)
- module (varchar(50))
- created_at, updated_at

### role_has_permissions

- role_id (UUID, FK -> roles.id)
- permission_id (UUID, FK -> permissions.id)
- created_at

### users

- id (UUID, PK)
- name (varchar(255))
- email (varchar(255), unique)
- password (varchar(255))
- avatar (varchar(500), nullable)
- role_id (UUID, FK -> roles.id, nullable)
- email_verified_at (timestamp, nullable)
- is_active (boolean, default true)
- last_login_at (timestamp, nullable)
- created_at, updated_at
- deleted_at (soft delete)

## Notes

- **UUID**: Tất cả tables đều dùng UUID làm primary key
- **Timestamps**: GORM tự động handle created_at & updated_at, không cần trigger
- **Soft Delete**: Users table có deleted_at cho soft delete
- **Foreign Keys**: ON DELETE CASCADE/SET NULL để maintain referential integrity
- **Indexes**: Đã tạo indexes cho các fields thường query (email, role_id, is_active)
