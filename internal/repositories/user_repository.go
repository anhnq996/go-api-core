package repository

import (
	model "anhnq/api-core/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository interface
type UserRepository interface {
	Create(user model.User) (model.User, error)
	FindAll() ([]model.User, error)
	FindByID(id string) (model.User, error)
	Update(id string, user model.User) (model.User, error)
	Delete(id string) error
	FindByEmail(email string) (*model.User, error)
}

// userRepository implement UserRepository với GORM
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository tạo user repository với GORM
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create tạo user mới
func (r *userRepository) Create(u model.User) (model.User, error) {
	// GORM tự động set ID nếu dùng UUID default
	if u.ID == "" {
		u.ID = uuid.New().String()
	}

	if err := r.db.Create(&u).Error; err != nil {
		return model.User{}, err
	}

	return u, nil
}

// FindAll lấy tất cả users
func (r *userRepository) FindAll() ([]model.User, error) {
	var users []model.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindByID tìm user theo ID
func (r *userRepository) FindByID(id string) (model.User, error) {
	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

// Update cập nhật user
func (r *userRepository) Update(id string, u model.User) (model.User, error) {
	var user model.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return model.User{}, err
	}

	// Update fields
	if err := r.db.Model(&user).Updates(u).Error; err != nil {
		return model.User{}, err
	}

	// Reload to get updated values
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return model.User{}, err
	}

	return user, nil
}

// Delete xóa user (soft delete)
func (r *userRepository) Delete(id string) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

// FindByEmail tìm user theo email
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
