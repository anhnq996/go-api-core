package auth

import (
	"context"
	"errors"
	"time"

	model "api-core/internal/models"
	repository "api-core/internal/repositories"
	"api-core/pkg/jwt"
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
	userRepo   repository.UserRepository
	jwtManager *jwt.Manager
	blacklist  *jwt.Blacklist
}

// NewService tạo auth service mới
func NewService(
	userRepo repository.UserRepository,
	jwtManager *jwt.Manager,
	blacklist *jwt.Blacklist,
) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		blacklist:  blacklist,
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
func (s *Service) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if !utils.CheckPassword(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Get user with role and permissions
	userWithRole, err := s.userRepo.GetUserWithRole(ctx, user.ID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Update last login
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Build response
	response := &LoginResponse{
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

	return response, nil
}

// RefreshToken làm mới access token
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Verify refresh token
	userIDStr, err := s.jwtManager.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Get user with role
	user, err := s.userRepo.GetUserWithRole(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
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
		return nil, err
	}

	// Build response
	response := &LoginResponse{
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

	return response, nil
}

// Logout đăng xuất (blacklist token)
func (s *Service) Logout(ctx context.Context, token string) error {
	// Get token expiry
	expiry, err := s.jwtManager.GetTokenExpiry(token)
	if err != nil {
		return err
	}

	// Add to blacklist
	return s.blacklist.Add(token, expiry)
}

// LogoutAll đăng xuất tất cả devices
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	// Blacklist all user tokens (7 days - max refresh token duration)
	expiry := utils.Now().Add(7 * 24 * time.Hour)
	return s.blacklist.AddUserTokens(userID.String(), expiry)
}

// GetUserInfo lấy thông tin user hiện tại
func (s *Service) GetUserInfo(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	// Get user with role
	user, err := s.userRepo.GetUserWithRole(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Get permissions
	var permissions []string
	if user.RoleID != nil {
		permissions, err = s.userRepo.GetUserPermissions(ctx, *user.RoleID)
		if err != nil {
			permissions = []string{}
		}
	}

	return &UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Avatar:      user.Avatar,
		Role:        buildRoleResponse(user.Role),
		Permissions: permissions,
	}, nil
}

// Register đăng ký user mới
func (s *Service) Register(ctx context.Context, name, email, password string, roleID *uuid.UUID) (*model.User, error) {
	// Check email exists
	_, err := s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		RoleID:   roleID,
		IsActive: true,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
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
