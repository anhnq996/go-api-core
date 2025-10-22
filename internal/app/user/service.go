package user

import (
	model "anhnq/api-core/internal/models"
	repository "anhnq/api-core/internal/repositories"
	"anhnq/api-core/pkg/cache"

	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service xử lý business logic cho user
type Service struct {
	repo  repository.UserRepository
	cache cache.Cache
}

const (
	cacheKeyAll = "users:all"
	cacheExpiry = 5 * time.Minute
)

// NewService tạo user service mới
func NewService(repo repository.UserRepository, cacheClient cache.Cache) *Service {
	return &Service{
		repo:  repo,
		cache: cacheClient,
	}
}

// GetAll lấy tất cả users
func (s *Service) GetAll() ([]model.User, error) {
	ctx := context.Background()

	// Try to get from cache first
	cached, err := s.cache.Remember(ctx, cacheKeyAll, cacheExpiry, func() (interface{}, error) {
		return s.repo.FindAll(ctx)
	})

	if err != nil {
		// If cache fails, try directly from DB
		return s.repo.FindAll(ctx)
	}

	// Convert cached data to []model.User
	users, ok := cached.([]model.User)
	if !ok {
		// Cache data invalid, fetch from DB
		return s.repo.FindAll(ctx)
	}

	return users, nil
}

// GetByID lấy user theo ID
func (s *Service) GetByID(id string) (*model.User, error) {
	ctx := context.Background()
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, userID)
}

// Create tạo user mới
func (s *Service) Create(user model.User) (*model.User, error) {
	ctx := context.Background()

	if err := s.repo.Create(ctx, &user); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cache.Del(ctx, cacheKeyAll)

	return &user, nil
}

// Update cập nhật user
func (s *Service) Update(id string, user model.User) (*model.User, error) {
	ctx := context.Background()
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, userID, &user); err != nil {
		return nil, err
	}

	// Get updated user
	updated, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cache.Del(ctx, cacheKeyAll, fmt.Sprintf("user:%s", id))

	return updated, nil
}

// Delete xóa user
func (s *Service) Delete(id string) error {
	ctx := context.Background()
	userID, err := uuid.Parse(id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, userID); err != nil {
		return err
	}

	// Invalidate cache
	s.cache.Del(ctx, cacheKeyAll, fmt.Sprintf("user:%s", id))

	return nil
}
