# Test Container Setup - PostgreSQL với Auto Migrate & Seeder

## 🎯 Tính năng

✅ **PostgreSQL Test Container** - Chạy PostgreSQL trong Docker container
✅ **Auto Migrate** - Tự động chạy migrations khi setup
✅ **Auto Seeder** - Tự động chạy seeders khi setup
✅ **Auto Cleanup** - Tự động cleanup database và container khi test xong

## 📝 Cách Sử Dụng

### 1. Basic Setup với Migrations và Seeders

```go
func TestMyFeature(t *testing.T) {
    // Setup test container với migrations và seeders
    config := SetupTestContainerConfig(t, true, true)
    defer CleanupTestContainerConfig(t, config)

    // Database đã sẵn sàng với:
    // - Tất cả tables từ migrations
    // - Tất cả data từ seeders
    // - Sẵn sàng để test
}
```

### 2. Setup chỉ với Migrations (không seed)

```go
func TestMyFeature(t *testing.T) {
    // Setup chỉ với migrations
    config := SetupTestContainerConfig(t, true, false)
    defer CleanupTestContainerConfig(t, config)

    // Database có tables nhưng không có seed data
}
```

### 3. Setup không Migrate/Seeder

```go
func TestMyFeature(t *testing.T) {
    // Setup container nhưng không migrate/seeder
    config := SetupTestContainerConfig(t, false, false)
    defer CleanupTestContainerConfig(t, config)

    // Database empty, có thể tự migrate
}
```

### 4. Manual Clean Database

```go
func TestWithManualClean(t *testing.T) {
    config := SetupTestContainerConfig(t, true, true)
    defer CleanupTestContainerConfig(t, config)

    // Create test data
    // ...

    // Clean database (xóa data nhưng giữ tables)
    CleanTestDBForContainer(t, config.DB)
}
```

### 5. Reset Database (Drop & Re-migrate)

```go
func TestWithReset(t *testing.T) {
    config := SetupTestContainerConfig(t, true, true)
    defer CleanupTestContainerConfig(t, config)

    // Reset database (drop tables và re-migrate)
    ResetTestDBForContainer(t, config.DB, true) // true = re-migrate

    // Tables đã được drop và re-migrate
}
```

## 🔧 Functions Available

### Setup Functions

- `SetupTestContainerConfig(t, enableMigrate, enableSeeder)` - Setup test container
  - `enableMigrate`: true để auto chạy migrations
  - `enableSeeder`: true để auto chạy seeders

### Cleanup Functions

- `CleanupTestContainerConfig(t, config)` - Auto cleanup:
  - Clean database data
  - Close database connection
  - Terminate container

### Manual Clean Functions

- `CleanTestDBForContainer(t, db)` - Clean data trong tables (giữ tables)
- `ResetTestDBForContainer(t, db, enableMigrate)` - Drop tables và re-migrate

## 📊 Test Container Details

- **Database**: PostgreSQL 16 Alpine
- **Database Name**: `test_db`
- **Username**: `test_user`
- **Password**: `test_password`
- **Port**: Auto-assigned by Docker
- **Connection**: Auto-managed by testcontainers

## ✅ Auto Cleanup Flow

```
Test Start
  ↓
Setup Container → Run Migrations → Run Seeders
  ↓
Test Runs
  ↓
Test Finishes
  ↓
Clean Database → Close Connection → Terminate Container
```

## 💡 Best Practices

1. **Always use defer**:

   ```go
   defer CleanupTestContainerConfig(t, config)
   ```

2. **Use enableMigrate=true** khi cần schema:

   ```go
   config := SetupTestContainerConfig(t, true, false)
   ```

3. **Use enableSeeder=true** khi cần test data:

   ```go
   config := SetupTestContainerConfig(t, true, true)
   ```

4. **Test isolation**: Mỗi test có container riêng

5. **Parallel tests**: Testcontainers supports parallel execution

## ⚠️ Requirements

- **Docker** phải chạy trên máy
- **testcontainers-go** package đã được cài
- **PostgreSQL migrations** ở `database/migrations/`
- **Seeders** ở `database/seeders/`

## 🚀 Example Test

```go
package test

import (
    "testing"
    "api-core/internal/repositories"
    "github.com/stretchr/testify/assert"
)

func TestUserRepositoryWithContainer(t *testing.T) {
    // Setup
    config := SetupTestContainerConfig(t, true, true)
    defer CleanupTestContainerConfig(t, config)

    // Create repository
    userRepo := repository.NewUserRepository(config.DB)

    // Test
    users, err := userRepo.FindAll(nil)
    assert.NoError(t, err)
    assert.NotNil(t, users)

    // Auto cleanup when test finishes
}
```

## 📈 Performance

- Container startup: ~3-5 seconds (first time), ~1-2 seconds (cached)
- Migrations: ~500ms
- Seeders: ~200ms
- Cleanup: ~100ms
- Container termination: ~500ms

**Total overhead**: ~5-8 seconds per test (first time), ~2-3 seconds (subsequent)

## 🔍 Troubleshooting

1. **Container không start**: Kiểm tra Docker đang chạy
2. **Migration fails**: Kiểm tra migrations ở `database/migrations/`
3. **Seeder fails**: Kiểm tra seeders ở `database/seeders/`
4. **Connection fails**: Kiểm tra container đã ready chưa

## 🎉 Benefits

✅ **Real Database**: Giống database production
✅ **Isolation**: Mỗi test có database riêng
✅ **Auto Setup**: Tự động migrate và seed
✅ **Auto Cleanup**: Tự động cleanup sau test
✅ **No Manual Setup**: Không cần setup database manually
