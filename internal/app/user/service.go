package user

import (
	model "anhnq/api-core/internal/models"
	repository "anhnq/api-core/internal/repositories"
	"anhnq/api-core/pkg/cache"
	"anhnq/api-core/pkg/storage"
	"anhnq/api-core/pkg/utils"

	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service xử lý business logic cho user
type Service struct {
	repo           repository.UserRepository
	cache          cache.Cache
	storageManager *storage.StorageManager
}

const (
	cacheKeyAll = "users:all"
	cacheExpiry = 5 * time.Minute
)

// NewService tạo user service mới
func NewService(repo repository.UserRepository, cacheClient cache.Cache, storageManager *storage.StorageManager) *Service {
	return &Service{
		repo:           repo,
		cache:          cacheClient,
		storageManager: storageManager,
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

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert avatar path to full URL
	s.convertAvatarToFullURL(user)

	return user, nil
}

// Create tạo user mới
func (s *Service) Create(user model.User, avatarFile *multipart.FileHeader) (*model.User, error) {
	ctx := context.Background()

	// Upload avatar nếu có
	if avatarFile != nil {
		uploadOptions := storage.GetImageUploadOptions(300, 300, 90) // 300x300, quality 90
		uploadOptions.Path = "avatars"                               // Store in avatars folder

		result, err := s.storageManager.UploadFile(ctx, avatarFile, uploadOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to upload avatar: %w", err)
		}

		user.Avatar = &result.Path
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		// Nếu tạo user thất bại, xóa avatar đã upload
		if user.Avatar != nil {
			s.storageManager.DeleteFile(ctx, *user.Avatar)
		}
		return nil, err
	}

	// Invalidate cache
	s.cache.Del(ctx, cacheKeyAll)

	// Convert avatar path to full URL
	s.convertAvatarToFullURL(&user)

	return &user, nil
}

// Update cập nhật user
func (s *Service) Update(id string, user model.User, avatarFile *multipart.FileHeader) (*model.User, error) {
	ctx := context.Background()
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	// Get current user để lấy avatar cũ
	currentUser, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Upload avatar mới nếu có
	if avatarFile != nil {
		uploadOptions := storage.GetImageUploadOptions(300, 300, 90) // 300x300, quality 90
		uploadOptions.Path = "avatars"                               // Store in avatars folder

		result, err := s.storageManager.UploadFile(ctx, avatarFile, uploadOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to upload avatar: %w", err)
		}

		user.Avatar = &result.Path
	}

	if err := s.repo.Update(ctx, userID, &user); err != nil {
		// Nếu update thất bại, xóa avatar mới đã upload
		if avatarFile != nil && user.Avatar != nil {
			s.storageManager.DeleteFile(ctx, *user.Avatar)
		}
		return nil, err
	}

	// Xóa avatar cũ nếu có avatar mới
	if avatarFile != nil && currentUser.Avatar != nil && *currentUser.Avatar != "" {
		if err := s.storageManager.DeleteFile(ctx, *currentUser.Avatar); err != nil {
			// Log error nhưng không fail operation
			fmt.Printf("Warning: Failed to delete old avatar: %v\n", err)
		}
	}

	// Get updated user
	updated, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Invalidate cache
	s.cache.Del(ctx, cacheKeyAll, fmt.Sprintf("user:%s", id))

	// Convert avatar path to full URL
	s.convertAvatarToFullURL(updated)

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

// GetListWithPagination lấy danh sách users với pagination, sort và search
func (s *Service) GetListWithPagination(page, perPage int, sort, order, search string) ([]model.User, *utils.Pagination, error) {
	ctx := context.Background()

	// Get users with pagination
	users, total, err := s.repo.FindAllWithPaginationAndRole(ctx, page, perPage, sort, order, search)
	if err != nil {
		return nil, nil, err
	}

	// Create pagination info
	pagination := utils.NewPagination(page, perPage, total)

	// Convert avatar paths to full URLs
	s.convertUsersAvatarToFullURL(users)

	return users, pagination, nil
}

// convertAvatarToFullURL converts avatar path to full URL
func (s *Service) convertAvatarToFullURL(user *model.User) {
	if user.Avatar != nil && *user.Avatar != "" {
		serverURL := os.Getenv("SERVER_URL")
		if serverURL == "" {
			serverURL = "http://localhost:3000"
		}

		storageURL := os.Getenv("STORAGE_LOCAL_URL")
		if storageURL == "" {
			storageURL = "/storages"
		}

		// Remove leading slash if exists
		avatarPath := strings.TrimPrefix(*user.Avatar, "/")

		// Create full URL using storage URL
		fullURL := fmt.Sprintf("%s%s/%s", strings.TrimSuffix(serverURL, "/"), storageURL, avatarPath)
		user.Avatar = &fullURL
	}
}

// convertUsersAvatarToFullURL converts avatar paths to full URLs for multiple users
func (s *Service) convertUsersAvatarToFullURL(users []model.User) {
	for i := range users {
		s.convertAvatarToFullURL(&users[i])
	}
}
