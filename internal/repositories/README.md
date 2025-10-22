# Repository Pattern

Repository layer với Generic Base Repository để tái sử dụng code.

## Pattern

### Base Repository

`BaseRepository[T]` cung cấp các CRUD operations cơ bản sử dụng Go generics:

```go
type Repository[T any] interface {
    Create(ctx context.Context, entity *T) error
    FindAll(ctx context.Context) ([]T, error)
    FindByID(ctx context.Context, id uuid.UUID) (*T, error)
    Update(ctx context.Context, id uuid.UUID, entity *T) error
    Delete(ctx context.Context, id uuid.UUID) error
    Count(ctx context.Context) (int64, error)
    Exists(ctx context.Context, id uuid.UUID) (bool, error)
    FindWhere(ctx context.Context, condition string, args ...interface{}) ([]T, error)
    FirstWhere(ctx context.Context, condition string, args ...interface{}) (*T, error)
    Paginate(ctx context.Context, page, perPage int) ([]T, int64, error)
    BulkCreate(ctx context.Context, entities []T) error
    // ...
}
```

### Extend Base Repository

Các repository cụ thể **embed** BaseRepository và chỉ thêm custom methods khi cần:

```go
// UserRepository extends base repository
type UserRepository interface {
    Repository[model.User] // Embed tất cả methods từ base

    // Chỉ thêm custom methods
    FindByEmail(ctx context.Context, email string) (*model.User, error)
    FindWithRole(ctx context.Context, id uuid.UUID) (*model.User, error)
}

// Implementation
type userRepository struct {
    *BaseRepository[model.User] // Embed implementation
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{
        BaseRepository: NewBaseRepository[model.User](db),
    }
}

// Chỉ implement custom methods
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
    // Sử dụng FirstWhere từ base repository
    return r.FirstWhere(ctx, "email = ?", email)
}
```

## Ưu điểm

✅ **DRY (Don't Repeat Yourself)** - Không cần viết lại CRUD cho mỗi repository
✅ **Type-safe** - Generics đảm bảo type safety
✅ **Flexible** - Có thể override methods khi cần
✅ **Consistent** - Tất cả repositories có interface giống nhau
✅ **Testable** - Dễ dàng mock base repository

## Usage Examples

### Example 1: UserRepository

```go
// Chỉ cần định nghĩa custom methods
type UserRepository interface {
    Repository[model.User]
    FindByEmail(ctx context.Context, email string) (*model.User, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{
        BaseRepository: NewBaseRepository[model.User](db),
    }
}

// Tự động có tất cả methods:
users, _ := userRepo.FindAll(ctx)                    // From base
user, _ := userRepo.FindByID(ctx, id)                // From base
userRepo.Create(ctx, &user)                          // From base
userRepo.Delete(ctx, id)                             // From base
userRepo.Paginate(ctx, 1, 10)                        // From base

// Custom methods
user, _ := userRepo.FindByEmail(ctx, "user@example.com")
```

### Example 2: ProductRepository

```go
// Tạo repository mới cho Product rất đơn giản
type ProductRepository interface {
    Repository[model.Product] // Đã có tất cả CRUD

    // Chỉ thêm custom methods nếu cần
    FindByCategory(ctx context.Context, category string) ([]model.Product, error)
    FindInStock(ctx context.Context) ([]model.Product, error)
}

type productRepository struct {
    *BaseRepository[model.Product]
}

func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{
        BaseRepository: NewBaseRepository[model.Product](db),
    }
}

// Custom method
func (r *productRepository) FindByCategory(ctx context.Context, category string) ([]model.Product, error) {
    return r.FindWhere(ctx, "category = ?", category)
}

func (r *productRepository) FindInStock(ctx context.Context) ([]model.Product, error) {
    return r.FindWhere(ctx, "stock > 0")
}

// Tất cả CRUD operations đã có sẵn từ BaseRepository!
```

### Example 3: Override Method

Nếu cần custom logic, override method:

```go
// Override FindAll để preload relationships
func (r *userRepository) FindAll(ctx context.Context) ([]model.User, error) {
    var users []model.User
    err := r.WithPreload("Role", "Profile").Find(&users).Error
    return users, err
}

// Override FindByID để thêm custom logic
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
    // Check cache first
    if cached := checkCache(id); cached != nil {
        return cached, nil
    }

    // Call base method
    user, err := r.BaseRepository.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Save to cache
    saveCache(user)

    return user, nil
}
```

## Available Methods

### CRUD Operations

- `Create(ctx, entity)` - Tạo mới
- `FindAll(ctx)` - Lấy tất cả
- `FindByID(ctx, id)` - Tìm theo ID
- `Update(ctx, id, entity)` - Cập nhật
- `Delete(ctx, id)` - Xóa (soft delete nếu model có DeletedAt)

### Query Methods

- `FindWhere(ctx, condition, args...)` - Tìm theo điều kiện
- `FirstWhere(ctx, condition, args...)` - Tìm 1 record
- `UpdateWhere(ctx, condition, updates, args...)` - Update theo điều kiện
- `DeleteWhere(ctx, condition, args...)` - Delete theo điều kiện

### Utility Methods

- `Count(ctx)` - Đếm tổng số
- `Exists(ctx, id)` - Kiểm tra tồn tại
- `Paginate(ctx, page, perPage)` - Phân trang
- `BulkCreate(ctx, entities)` - Tạo nhiều records

### Database Access

- `DB()` - Truy cập GORM DB instance
- `WithPreload(associations...)` - Preload relationships
- `Transaction(fn)` - Run in transaction

## Best Practices

### 1. Chỉ thêm custom methods khi thật sự cần

```go
// ✅ Good - Chỉ custom methods
type UserRepository interface {
    Repository[model.User]
    FindByEmail(ctx context.Context, email string) (*model.User, error)
}

// ❌ Bad - Duplicate base methods
type UserRepository interface {
    Repository[model.User]
    FindAll(ctx context.Context) ([]model.User, error) // Đã có trong base!
    FindByEmail(ctx context.Context, email string) (*model.User, error)
}
```

### 2. Sử dụng FirstWhere/FindWhere thay vì viết raw query

```go
// ✅ Good - Sử dụng base methods
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
    return r.FirstWhere(ctx, "email = ?", email)
}

// ❌ Avoid - Viết lại logic
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
    var user model.User
    err := r.db.Where("email = ?", email).First(&user).Error
    // ...
}
```

### 3. Override khi cần thêm logic phức tạp

```go
// Override FindAll để preload role mặc định
func (r *userRepository) FindAll(ctx context.Context) ([]model.User, error) {
    var users []model.User
    err := r.WithPreload("Role").Find(&users).Error
    return users, err
}
```

### 4. Luôn truyền context

```go
// ✅ Good
users, err := repo.FindAll(ctx)

// ❌ Bad
users, err := repo.FindAll(context.Background()) // Hardcoded context
```

## Migration từ Old Pattern

### Trước:

```go
// Phải viết lại tất cả methods
type UserRepository interface {
    Create(user model.User) (model.User, error)
    FindAll() ([]model.User, error)
    FindByID(id string) (model.User, error)
    Update(id string, user model.User) (model.User, error)
    Delete(id string) error
}

func (r *userRepository) Create(u model.User) (model.User, error) {
    if err := r.db.Create(&u).Error; err != nil {
        return model.User{}, err
    }
    return u, nil
}

func (r *userRepository) FindAll() ([]model.User, error) {
    var users []model.User
    if err := r.db.Find(&users).Error; err != nil {
        return nil, err
    }
    return users, nil
}

// ... phải viết hết tất cả methods
```

### Sau:

```go
// Chỉ cần định nghĩa custom methods
type UserRepository interface {
    Repository[model.User] // Inherit tất cả CRUD
    FindByEmail(ctx context.Context, email string) (*model.User, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{
        BaseRepository: NewBaseRepository[model.User](db),
    }
}

// Chỉ implement custom methods
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
    return r.FirstWhere(ctx, "email = ?", email)
}

// Tất cả CRUD đã có sẵn từ BaseRepository!
```

## Testing

```go
func TestUserRepository(t *testing.T) {
    db := setupTestDB()
    repo := NewUserRepository(db)

    // Test base methods
    users, err := repo.FindAll(context.Background())
    assert.NoError(t, err)

    user, err := repo.FindByID(context.Background(), uuid.New())
    assert.NoError(t, err)

    // Test custom methods
    user, err = repo.FindByEmail(context.Background(), "test@example.com")
    assert.NoError(t, err)
}
```

## Tạo Repository Mới

1. Định nghĩa interface (embed Repository[T])
2. Tạo struct (embed \*BaseRepository[T])
3. Implement constructor
4. Thêm custom methods (nếu cần)

```go
// 1. Interface
type OrderRepository interface {
    Repository[model.Order]
    FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Order, error)
}

// 2. Struct
type orderRepository struct {
    *BaseRepository[model.Order]
}

// 3. Constructor
func NewOrderRepository(db *gorm.DB) OrderRepository {
    return &orderRepository{
        BaseRepository: NewBaseRepository[model.Order](db),
    }
}

// 4. Custom methods
func (r *orderRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Order, error) {
    return r.FindWhere(ctx, "user_id = ?", userID)
}

// Done! Đã có repository đầy đủ CRUD + custom methods
```

## Summary

- **BaseRepository**: Generic repository với tất cả CRUD operations
- **Extend**: Các repository cụ thể embed BaseRepository
- **Custom**: Chỉ thêm methods đặc thù cho từng entity
- **Override**: Override khi cần custom logic
- **Reusable**: Code gọn, tái sử dụng cao, dễ maintain
