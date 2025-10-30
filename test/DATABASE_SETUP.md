## Database Testing Setup

### ✅ Đã Setup

1. **SQLite Driver không cần CGO**: Sử dụng `github.com/glebarez/sqlite` (pure Go)
2. **In-memory Database**: Mỗi test sử dụng database riêng biệt
3. **Auto Cleanup**: Database tự động được clean sau mỗi test

### 📝 Cách Sử Dụng

#### 1. Basic Test (không cần models)

```go
func TestMyFeature(t *testing.T) {
    // Setup test config
    config := SetupTestConfig(t)
    defer CleanupTestConfig(t, config)

    // Database đã sẵn sàng để sử dụng
    // config.DB là *gorm.DB instance
}
```

#### 2. Test với Database Migrations

```go
import (
    model "api-core/internal/models"
)

func TestWithMigrations(t *testing.T) {
    // Setup với migrations
    config := SetupTestConfigWithDB(t, &model.User{})
    defer CleanupTestConfig(t, config)

    // Models đã được migrate
    // Sẵn sàng để test
}
```

#### 3. Manual Clean/Reset Database

```go
func TestWithManualClean(t *testing.T) {
    config := SetupTestConfigWithDB(t, &model.User{})
    defer CleanupTestConfig(t, config)

    // Tạo test data
    // ...

    // Clean database (xóa tất cả data nhưng giữ tables)
    CleanTestDB(t, config.DB)

    // Hoặc reset database (drop tables và re-migrate)
    ResetTestDB(t, config.DB, &model.User{})
}
```

### 🔧 Available Functions

#### Setup Functions

- `SetupTestConfig(t)` - Setup config với database in-memory
- `SetupTestConfigWithDB(t, models...)` - Setup với migrations
- `SetupTestDB(t, db, models...)` - Migrate models vào database

#### Cleanup Functions

- `CleanupTestConfig(t, config)` - Cleanup resources (auto clean database)
- `CleanTestDB(t, db)` - Clean tất cả data trong tables
- `ResetTestDB(t, db, models...)` - Drop tables và re-migrate

### 💡 Best Practices

1. **Always use defer**:

   ```go
   defer CleanupTestConfig(t, config)
   ```

2. **Use in-memory database**: Tự động, không cần config

3. **Clean between tests**:

   - Auto cleanup trong `defer`
   - Hoặc manual clean nếu cần

4. **Test isolation**: Mỗi test có database riêng

### ⚠️ Lưu Ý

- User model có thể cần adjust để compatible với SQLite
- Nếu migrate fail, check model definition
- Database là in-memory nên tự động reset khi test kết thúc

### ✅ Test Status

- ✅ Database connection works (no CGO required)
- ✅ Clean/Reset functions ready
- ✅ Auto cleanup in defer
- ⚠️ Some models may need SQLite compatibility adjustments
