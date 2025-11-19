package user

import (
	model "api-core/internal/models"
	repository "api-core/internal/repositories"
	"api-core/pkg/cache"
	"api-core/pkg/fcm"
	"api-core/pkg/i18n"
	"api-core/pkg/logger"
	"api-core/pkg/response"
	"api-core/pkg/storage"
	"api-core/pkg/utils"

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
	fcmClient      *fcm.Client // Optional: nil nếu FCM chưa được cấu hình
}

const (
	cacheKeyAll = "users:all"
	cacheExpiry = 5 * time.Minute
)

// NewService tạo user service mới
func NewService(
	repo repository.UserRepository,
	cacheClient cache.Cache,
	storageManager *storage.StorageManager,
	fcmClient *fcm.Client, // Optional: có thể nil
) *Service {
	return &Service{
		repo:           repo,
		cache:          cacheClient,
		storageManager: storageManager,
		fcmClient:      fcmClient,
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
func (s *Service) GetByID(ctx context.Context, id string) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)
	userID, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequestResponse(lang, response.CodeInvalidInput, nil)
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return response.NotFoundResponse(lang, response.CodeUserNotFound)
	}

	// Convert avatar path to full URL
	s.convertAvatarToFullURL(user)

	return response.SuccessResponse(lang, response.CodeSuccess, user)
}

// Create tạo user mới (có thể nhận FCM token để gửi notification)
func (s *Service) Create(ctx context.Context, user model.User, avatarFile *multipart.FileHeader, fcmToken ...string) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Upload avatar nếu có
	if avatarFile != nil {
		uploadOptions := storage.GetImageUploadOptions(300, 300, 90) // 300x300, quality 90
		uploadOptions.Path = "avatars"                               // Store in avatars folder

		result, err := s.storageManager.UploadFile(ctx, avatarFile, uploadOptions)
		if err != nil {
			return response.InternalServerErrorResponse(lang, response.CodeFileUploadFailed)
		}

		user.Avatar = &result.Path
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		// Nếu tạo user thất bại, xóa avatar đã upload
		if user.Avatar != nil {
			s.storageManager.DeleteFile(ctx, *user.Avatar)
		}
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	// Invalidate cache
	s.cache.Del(ctx, cacheKeyAll)

	// Convert avatar path to full URL
	s.convertAvatarToFullURL(&user)

	// Gửi FCM notification chào mừng user mới (background, không block response)
	var token string
	if len(fcmToken) > 0 && fcmToken[0] != "" {
		token = fcmToken[0]
	}
	go s.sendWelcomeNotification(context.Background(), &user, token)

	return response.SuccessResponse(lang, response.CodeCreated, user)
}

// Update cập nhật user
func (s *Service) Update(ctx context.Context, id string, user model.User, avatarFile *multipart.FileHeader) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)
	userID, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequestResponse(lang, response.CodeInvalidInput, nil)
	}

	// Get current user để lấy avatar cũ
	currentUser, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return response.NotFoundResponse(lang, response.CodeUserNotFound)
	}

	// Upload avatar mới nếu có
	if avatarFile != nil {
		uploadOptions := storage.GetImageUploadOptions(300, 300, 90) // 300x300, quality 90
		uploadOptions.Path = "avatars"                               // Store in avatars folder

		result, err := s.storageManager.UploadFile(ctx, avatarFile, uploadOptions)
		if err != nil {
			return response.InternalServerErrorResponse(lang, response.CodeFileUploadFailed)
		}

		user.Avatar = &result.Path
	}

	if err := s.repo.Update(ctx, userID, &user); err != nil {
		// Nếu update thất bại, xóa avatar mới đã upload
		if avatarFile != nil && user.Avatar != nil {
			s.storageManager.DeleteFile(ctx, *user.Avatar)
		}
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
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
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	// Invalidate cache
	s.cache.Del(ctx, cacheKeyAll, fmt.Sprintf("user:%s", id))

	// Convert avatar path to full URL
	s.convertAvatarToFullURL(updated)

	return response.SuccessResponse(lang, response.CodeUpdated, updated)
}

// Delete xóa user
func (s *Service) Delete(ctx context.Context, id string) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)
	userID, err := uuid.Parse(id)
	if err != nil {
		return response.BadRequestResponse(lang, response.CodeInvalidInput, nil)
	}

	if err := s.repo.Delete(ctx, userID); err != nil {
		return response.NotFoundResponse(lang, response.CodeUserNotFound)
	}

	// Invalidate cache
	s.cache.Del(ctx, cacheKeyAll, fmt.Sprintf("user:%s", id))

	return response.SuccessResponse(lang, response.CodeDeleted, nil)
}

// GetListWithPagination lấy danh sách users với pagination, sort và search
func (s *Service) GetListWithPagination(ctx context.Context, page, perPage int, sort, order, search string) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Get users with pagination
	users, total, err := s.repo.FindAllWithPaginationAndRole(ctx, page, perPage, sort, order, search)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	// Create pagination info
	pagination := utils.NewPagination(page, perPage, total)

	// Convert avatar paths to full URLs
	s.convertUsersAvatarToFullURL(users)

	// Create response data
	responseData := utils.PaginatedResponse(users, pagination)
	meta := &response.Meta{
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      pagination.Total,
		TotalPages: pagination.TotalPages,
	}

	return response.SuccessResponseWithMeta(lang, response.CodeSuccess, responseData, meta)
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

// SendNotification gửi FCM notification đến user
// Ví dụ sử dụng: gửi notification khi user được tạo mới, cập nhật, etc.
// DEPRECATED: Sử dụng SendNotificationToToken hoặc SendNotificationToUser thay thế
func (s *Service) SendNotification(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error {
	return s.SendNotificationToUser(ctx, userID, title, body, data)
}

// SendNotificationToToken gửi FCM notification đến một token cụ thể
func (s *Service) SendNotificationToToken(ctx context.Context, token string, title, body string, data map[string]string) (string, error) {
	if s.fcmClient == nil {
		return "", fmt.Errorf("FCM client chưa được khởi tạo")
	}

	if token == "" {
		return "", fmt.Errorf("FCM token không được để trống")
	}

	// Tạo notification
	notification := fcm.NewNotificationBuilder().
		SetTitle(title).
		SetBody(body).
		Build()

	// Gửi notification
	messageID, err := s.fcmClient.SendToToken(ctx, token, notification, data)
	if err != nil {
		return "", fmt.Errorf("failed to send FCM notification: %w", err)
	}

	return messageID, nil
}

// SendNotificationToUser gửi FCM notification đến user (cần có FCM token trong DB)
// Cần implement getUserFCMToken để lấy token từ database
func (s *Service) SendNotificationToUser(ctx context.Context, userID uuid.UUID, title, body string, data map[string]string) error {
	if s.fcmClient == nil {
		return fmt.Errorf("FCM client chưa được khởi tạo")
	}

	// TODO: Lấy FCM token từ database
	// Ví dụ:
	// user, err := s.repo.FindByID(ctx, userID)
	// if err != nil {
	// 	return fmt.Errorf("user not found: %w", err)
	// }
	// if user.FCMToken == nil || *user.FCMToken == "" {
	// 	return fmt.Errorf("user does not have FCM token")
	// }
	// token := *user.FCMToken

	// Tạm thời return error để nhắc implement
	return fmt.Errorf("cần implement getUserFCMToken() để lấy token từ database")
}

// SendNotificationToTokens gửi FCM notification đến nhiều tokens (multicast)
func (s *Service) SendNotificationToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) (int, int, error) {
	if s.fcmClient == nil {
		return 0, 0, fmt.Errorf("FCM client chưa được khởi tạo")
	}

	if len(tokens) == 0 {
		return 0, 0, fmt.Errorf("danh sách tokens không được để trống")
	}

	// Tạo notification
	notification := fcm.NewNotificationBuilder().
		SetTitle(title).
		SetBody(body).
		Build()

	// Gửi notification
	response, err := s.fcmClient.SendToTokens(ctx, tokens, notification, data)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to send FCM notifications: %w", err)
	}

	return response.SuccessCount, response.FailureCount, nil
}

// sendWelcomeNotification gửi notification chào mừng user mới (background)
func (s *Service) sendWelcomeNotification(ctx context.Context, user *model.User, fcmToken string) {
	// Nếu không có FCM client hoặc không có token, bỏ qua
	if s.fcmClient == nil {
		return
	}

	// Nếu không có token, không gửi notification
	if fcmToken == "" {
		// Không log nếu không có token (bình thường khi client chưa cung cấp token)
		return
	}

	// Tạo notification
	notification := fcm.NewNotificationBuilder().
		SetTitle("Chào mừng đến với ApiCore!").
		SetBody(fmt.Sprintf("Xin chào %s! Tài khoản của bạn đã được tạo thành công.", user.Name)).
		Build()

	// Prepare data
	data := map[string]string{
		"type":      "user_created",
		"user_id":   user.ID.String(),
		"email":     user.Email,
		"action":    "view_profile",
		"deep_link": fmt.Sprintf("app://users/%s", user.ID),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Gửi notification trong goroutine riêng để có context timeout riêng
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		messageID, err := s.fcmClient.SendToToken(ctx, fcmToken, notification, data)
		if err != nil {
			// Log error nhưng không fail operation
			logger.Errorf("Failed to send welcome notification to user %s: %v", user.ID, err)
			return
		}

		logger.Infof("Welcome notification sent to user %s: message_id=%s", user.ID, messageID)
	}()
}
