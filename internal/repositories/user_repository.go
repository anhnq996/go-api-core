package repository

import (
	model "anhnq/api-core/internal/models"
	"errors"
	"time"

	"github.com/google/uuid"
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

// userRepository implement UserRepository
type userRepository struct {
	base *BaseRepository[model.User]
}

func NewUserRepository() UserRepository {
	return &userRepository{
		base: NewBaseRepository[model.User](),
	}
}

// Create
func (r *userRepository) Create(u model.User) (model.User, error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return r.base.Create(u.ID, u)
}

func (r *userRepository) FindAll() ([]model.User, error) {
	return r.base.FindAll()
}

func (r *userRepository) FindByID(id string) (model.User, error) {
	return r.base.FindByID(id)
}

func (r *userRepository) Update(id string, u model.User) (model.User, error) {
	u.UpdatedAt = time.Now()
	return r.base.Update(id, u)
}

func (r *userRepository) Delete(id string) error {
	return r.base.Delete(id)
}

// Custom method riêng cho UserRepository
func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	users, _ := r.base.FindAll()
	for _, u := range users {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, errors.New("user not found")
}
