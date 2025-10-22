package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	model "anhnq/api-core/internal/models"
	repository "anhnq/api-core/internal/repositories"
	"anhnq/api-core/pkg/cache"
)

type Service struct {
	repo  repository.UserRepository
	cache cache.Cache
}

func NewService(r repository.UserRepository, c cache.Cache) *Service {
	return &Service{
		repo:  r,
		cache: c,
	}
}

func (s *Service) GetAll() ([]model.User, error) {
	ctx := context.Background()
	cacheKey := "users:all"

	// Try get from cache using Remember pattern
	result, err := s.cache.Remember(ctx, cacheKey, 5*time.Minute, func() (interface{}, error) {
		// Cache miss - fetch from database
		return s.repo.FindAll()
	})

	if err != nil {
		// Cache error - fallback to database
		return s.repo.FindAll()
	}

	// Parse result
	var users []model.User
	jsonData, _ := json.Marshal(result)
	if err := json.Unmarshal(jsonData, &users); err != nil {
		// Parse error - fallback to database
		return s.repo.FindAll()
	}

	return users, nil
}

func (s *Service) GetByID(id string) (model.User, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%s", id)

	// Remember pattern for single user
	result, err := s.cache.Remember(ctx, cacheKey, 10*time.Minute, func() (interface{}, error) {
		return s.repo.FindByID(id)
	})

	if err != nil {
		return s.repo.FindByID(id)
	}

	// Parse result
	var user model.User
	jsonData, _ := json.Marshal(result)
	if err := json.Unmarshal(jsonData, &user); err != nil {
		return s.repo.FindByID(id)
	}

	return user, nil
}

func (s *Service) Create(u model.User) (model.User, error) {
	created, err := s.repo.Create(u)
	if err != nil {
		return model.User{}, err
	}

	// Invalidate cache list
	ctx := context.Background()
	s.cache.Del(ctx, "users:all")

	return created, nil
}

func (s *Service) Update(id string, u model.User) (model.User, error) {
	updated, err := s.repo.Update(id, u)
	if err != nil {
		return model.User{}, err
	}

	// Invalidate caches
	ctx := context.Background()
	s.cache.Del(ctx, "users:all", fmt.Sprintf("user:%s", id))

	return updated, nil
}

func (s *Service) Delete(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}

	// Invalidate caches
	ctx := context.Background()
	s.cache.Del(ctx, "users:all", fmt.Sprintf("user:%s", id))

	return nil
}
