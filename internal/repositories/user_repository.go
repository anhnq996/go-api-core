package repository

import (
	"context"
	"time"

	model "api-core/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository interface extends base repository với custom methods
type UserRepository interface {
	Repository[model.User] // Embed base repository interface

	// User management methods
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindWithRole(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindAllWithRole(ctx context.Context) ([]model.User, error)
	FindAllWithPaginationAndRole(ctx context.Context, page, perPage int, sort, order, search string) ([]model.User, int64, error)

	// Auth-related methods (moved from AuthRepository)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserWithRole(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
}

// userRepository implementation
type userRepository struct {
	*BaseRepository[model.User] // Embed base repository
}

// NewUserRepository tạo user repository mới
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository[model.User](db, true), // Enable action events for UserRepository
	}
}

// FindByEmail tìm user theo email (custom method)
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.FirstWhere(ctx, "email = ? AND is_active = ?", email, true)
}

// FindWithRole tìm user kèm role (custom method)
func (r *userRepository) FindWithRole(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.WithPreload("Role").
		Where("id = ? AND is_active = ?", id, true).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAllWithRole lấy tất cả users kèm role (custom method)
func (r *userRepository) FindAllWithRole(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.WithPreload("Role").Find(&users).Error
	return users, err
}

// Override FindAll để preload role by default
func (r *userRepository) FindAll(ctx context.Context) ([]model.User, error) {
	return r.FindAllWithRole(ctx)
}

// Override FindByID để preload role by default
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return r.FindWithRole(ctx, id)
}

// FindAllWithPaginationAndRole lấy users với pagination, sort, search và preload role
func (r *userRepository) FindAllWithPaginationAndRole(ctx context.Context, page, perPage int, sort, order, search string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// Set defaults
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}
	if order == "" {
		order = "asc"
	}
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// Build query
	query := r.db.WithContext(ctx).Model(&model.User{})

	// Add search condition
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Add sorting (chỉ sort nếu có truyền sort field)
	if sort != "" {
		sortField := sort
		if order == "desc" {
			sortField = sort + " DESC"
		}
		query = query.Order(sortField)
	}

	// Add pagination and execute with preload
	offset := (page - 1) * perPage
	err := query.Preload("Role").
		Offset(offset).
		Limit(perPage).
		Find(&users).Error

	return users, total, err
}

// Auth-related methods implementation (moved from AuthRepository)

// GetUserByEmail lấy user theo email (alias cho FindByEmail)
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.FindByEmail(ctx, email)
}

// GetUserWithRole lấy user kèm role và permissions (alias cho FindWithRole)
func (r *userRepository) GetUserWithRole(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.DB().WithContext(ctx).
		Preload("Role").
		Preload("Role.Permissions").
		Where("id = ? AND is_active = ?", id, true).
		First(&user).Error

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserPermissions lấy danh sách permissions của user qua role
func (r *userRepository) GetUserPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	var permissions []model.Permission

	err := r.DB().WithContext(ctx).
		Table("permissions").
		Joins("INNER JOIN role_has_permissions ON permissions.id = role_has_permissions.permission_id").
		Where("role_has_permissions.role_id = ?", roleID).
		Find(&permissions).Error

	if err != nil {
		return nil, err
	}

	// Convert to string array
	permissionNames := make([]string, len(permissions))
	for i, p := range permissions {
		permissionNames[i] = p.Name
	}

	return permissionNames, nil
}

// UpdateLastLogin cập nhật thời gian login cuối
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.UpdateWhere(ctx, "id = ?", map[string]interface{}{
		"last_login_at": now,
	}, userID)
}
