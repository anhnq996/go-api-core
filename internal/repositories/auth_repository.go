package repository

import (
	"context"
	"time"

	model "anhnq/api-core/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthRepository interface - extend base repository
type AuthRepository interface {
	Repository[model.User] // Embed base repository interface

	// Custom methods cho authentication
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserWithRole(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetUserPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
}

// authRepository implementation
type authRepository struct {
	*BaseRepository[model.User] // Embed base repository
}

// NewAuthRepository tạo auth repository mới
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{
		BaseRepository: NewBaseRepository[model.User](db),
	}
}

// GetUserByEmail lấy user theo email (custom method)
func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.FirstWhere(ctx, "email = ? AND is_active = ?", email, true)
}

// GetUserWithRole lấy user kèm role và permissions (custom method)
func (r *authRepository) GetUserWithRole(ctx context.Context, id uuid.UUID) (*model.User, error) {
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

// GetUserPermissions lấy danh sách permissions của user qua role (custom method)
func (r *authRepository) GetUserPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
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

// UpdateLastLogin cập nhật thời gian login cuối (custom method)
func (r *authRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.UpdateWhere(ctx, "id = ?", map[string]interface{}{
		"last_login_at": now,
	}, userID)
}
