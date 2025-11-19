package auth

import (
	"context"
	"errors"
	"mime/multipart"
	"time"

	model "api-core/internal/models"
	repository "api-core/internal/repositories"
	"api-core/pkg/i18n"
	"api-core/pkg/jwt"
	"api-core/pkg/response"
	"api-core/pkg/storage"
	"api-core/pkg/utils"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
)

// Service xử lý business logic cho auth
type Service struct {
	userRepo       repository.UserRepository
	jwtManager     *jwt.Manager
	blacklist      *jwt.Blacklist
	storageManager *storage.StorageManager
}

// NewService tạo auth service mới
func NewService(
	userRepo repository.UserRepository,
	jwtManager *jwt.Manager,
	blacklist *jwt.Blacklist,
	storageManager *storage.StorageManager,
) *Service {
	return &Service{
		userRepo:       userRepo,
		jwtManager:     jwtManager,
		blacklist:      blacklist,
		storageManager: storageManager,
	}
}

// LoginResponse response cho login
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresAt    string        `json:"expires_at"`
	TokenType    string        `json:"token_type"`
}

// UserResponse user info trong response
type UserResponse struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Email       string        `json:"email"`
	Avatar      *string       `json:"avatar"`
	Role        *RoleResponse `json:"role"`
	Permissions []string      `json:"permissions"`
}

// RoleResponse role info trong response
type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
}

// Login xử lý login
func (s *Service) Login(ctx context.Context, email, password string) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return response.UnauthorizedResponse(lang, response.CodeInvalidCredentials)
	}

	// Check if user is active
	if !user.IsActive {
		return response.ForbiddenResponse(lang, response.CodeAccountDisabled)
	}

	// Verify password
	if !utils.CheckPassword(password, user.Password) {
		return response.UnauthorizedResponse(lang, response.CodeInvalidCredentials)
	}

	// Get user with role and permissions
	userWithRole, err := s.userRepo.GetUserWithRole(ctx, user.ID)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	// Get permissions
	var permissions []string
	if userWithRole.RoleID != nil {
		permissions, err = s.userRepo.GetUserPermissions(ctx, *userWithRole.RoleID)
		if err != nil {
			permissions = []string{}
		}
	}

	// Generate JWT tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		getRoleName(userWithRole.Role),
		map[string]interface{}{
			"name": user.Name,
		},
	)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	// Update last login
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Build response
	loginResp := &LoginResponse{
		User: &UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			Avatar:      user.Avatar,
			Role:        buildRoleResponse(userWithRole.Role),
			Permissions: permissions,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		TokenType:    tokenPair.TokenType,
	}

	return response.SuccessResponse(lang, response.CodeLoginSuccess, loginResp)
}

// RefreshToken làm mới access token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Verify refresh token
	userIDStr, err := s.jwtManager.VerifyRefreshToken(refreshToken)
	if err != nil {
		if err == jwt.ErrExpiredToken {
			return response.UnauthorizedResponse(lang, response.CodeTokenExpired)
		}
		if err == jwt.ErrInvalidToken {
			return response.UnauthorizedResponse(lang, response.CodeTokenInvalid)
		}
		return response.UnauthorizedResponse(lang, response.CodeTokenInvalid)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return response.UnauthorizedResponse(lang, response.CodeInvalidCredentials)
	}

	// Get user with role
	user, err := s.userRepo.GetUserWithRole(ctx, userID)
	if err != nil {
		return response.NotFoundResponse(lang, response.CodeUserNotFound)
	}

	// Check if user is active
	if !user.IsActive {
		return response.ForbiddenResponse(lang, response.CodeAccountDisabled)
	}

	// Get permissions
	var permissions []string
	if user.RoleID != nil {
		permissions, err = s.userRepo.GetUserPermissions(ctx, *user.RoleID)
		if err != nil {
			permissions = []string{}
		}
	}

	// Generate new tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(
		user.ID.String(),
		user.Email,
		getRoleName(user.Role),
		map[string]interface{}{
			"name": user.Name,
		},
	)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	// Build response
	loginResp := &LoginResponse{
		User: &UserResponse{
			ID:          user.ID,
			Name:        user.Name,
			Email:       user.Email,
			Avatar:      user.Avatar,
			Role:        buildRoleResponse(user.Role),
			Permissions: permissions,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		TokenType:    tokenPair.TokenType,
	}

	return response.SuccessResponse(lang, response.CodeTokenRefreshed, loginResp)
}

// Logout đăng xuất (blacklist token)
func (s *Service) Logout(ctx context.Context, token string) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Get token expiry
	expiry, err := s.jwtManager.GetTokenExpiry(token)
	if err != nil {
		return response.UnauthorizedResponse(lang, response.CodeTokenInvalid)
	}

	// Add to blacklist
	if err := s.blacklist.Add(token, expiry); err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	return response.SuccessResponse(lang, response.CodeLogoutSuccess, nil)
}

// LogoutAll đăng xuất tất cả devices
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Blacklist all user tokens (7 days - max refresh token duration)
	expiry := utils.Now().Add(7 * 24 * time.Hour)
	if err := s.blacklist.AddUserTokens(userID.String(), expiry); err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	return response.SuccessResponse(lang, response.CodeLogoutSuccess, nil)
}

// GetUserInfo lấy thông tin user hiện tại
func (s *Service) GetUserInfo(ctx context.Context, userID uuid.UUID) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Get user with role
	user, err := s.userRepo.GetUserWithRole(ctx, userID)
	if err != nil {
		return response.NotFoundResponse(lang, response.CodeUserNotFound)
	}

	// Get permissions
	var permissions []string
	if user.RoleID != nil {
		permissions, err = s.userRepo.GetUserPermissions(ctx, *user.RoleID)
		if err != nil {
			permissions = []string{}
		}
	}

	userResp := &UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Avatar:      user.Avatar,
		Role:        buildRoleResponse(user.Role),
		Permissions: permissions,
	}

	return response.SuccessResponse(lang, response.CodeSuccess, userResp)
}

// Register đăng ký user mới
func (s *Service) Register(ctx context.Context, name, email, password string, roleID *uuid.UUID, avatarFile *multipart.FileHeader) *response.Response {
	lang := i18n.GetLanguageFromContext(ctx)

	// Check email exists
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return response.ConflictResponse(lang, response.CodeEmailAlreadyExists)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	// Create user
	user := &model.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		RoleID:   roleID,
		IsActive: true,
	}

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

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		// Nếu tạo user thất bại, xóa avatar đã upload
		if user.Avatar != nil {
			s.storageManager.DeleteFile(ctx, *user.Avatar)
		}
		return response.InternalServerErrorResponse(lang, response.CodeInternalServerError)
	}

	return response.SuccessResponse(lang, response.CodeCreated, user)
}

// Helper functions

func getRoleName(role *model.Role) string {
	if role == nil {
		return "user"
	}
	return role.Name
}

func buildRoleResponse(role *model.Role) *RoleResponse {
	if role == nil {
		return nil
	}
	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
	}
}
