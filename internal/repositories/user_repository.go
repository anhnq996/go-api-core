package repository

import (
	"context"

	model "anhnq/api-core/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository interface extends base repository với custom methods
type UserRepository interface {
	Repository[model.User] // Embed base repository interface

	// Custom methods specific cho User
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindWithRole(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindAllWithRole(ctx context.Context) ([]model.User, error)
}

// userRepository implementation
type userRepository struct {
	*BaseRepository[model.User] // Embed base repository
}

// NewUserRepository tạo user repository mới
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository[model.User](db),
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
