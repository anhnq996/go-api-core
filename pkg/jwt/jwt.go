package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config cấu hình cho JWT
type Config struct {
	SecretKey            string        // Secret key để sign token
	AccessTokenDuration  time.Duration // Thời gian hết hạn access token (default: 15 phút)
	RefreshTokenDuration time.Duration // Thời gian hết hạn refresh token (default: 7 ngày)
	Issuer               string        // Issuer của token (default: "apicore")
}

// Claims chứa thông tin trong JWT token
type Claims struct {
	UserID   string                 `json:"user_id"`
	Email    string                 `json:"email"`
	Role     string                 `json:"role"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// TokenPair chứa cả access token và refresh token
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// Manager quản lý JWT tokens
type Manager struct {
	config Config
}

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrTokenNotFound    = errors.New("token not found")
)

// NewManager tạo JWT manager mới
func NewManager(config Config) *Manager {
	// Set defaults
	if config.AccessTokenDuration == 0 {
		config.AccessTokenDuration = 15 * time.Minute
	}
	if config.RefreshTokenDuration == 0 {
		config.RefreshTokenDuration = 7 * 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "apicore"
	}

	return &Manager{
		config: config,
	}
}

// GenerateToken tạo access token
func (m *Manager) GenerateToken(userID, email, role string, metadata map[string]interface{}) (string, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.AccessTokenDuration)

	claims := Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		Metadata: metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.SecretKey))
}

// GenerateRefreshToken tạo refresh token
func (m *Manager) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(m.config.RefreshTokenDuration)

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		Issuer:    m.config.Issuer,
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.SecretKey))
}

// GenerateTokenPair tạo cả access token và refresh token
func (m *Manager) GenerateTokenPair(userID, email, role string, metadata map[string]interface{}) (*TokenPair, error) {
	accessToken, err := m.GenerateToken(userID, email, role, metadata)
	if err != nil {
		return nil, err
	}

	refreshToken, err := m.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(m.config.AccessTokenDuration),
		TokenType:    "Bearer",
	}, nil
}

// VerifyToken xác thực và parse token
func (m *Manager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Kiểm tra signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// VerifyRefreshToken xác thực refresh token
func (m *Manager) VerifyRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignature
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", ErrExpiredToken
		}
		return "", ErrInvalidToken
	}

	if !token.Valid {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}

// RefreshAccessToken tạo access token mới từ refresh token
func (m *Manager) RefreshAccessToken(refreshToken, email, role string, metadata map[string]interface{}) (*TokenPair, error) {
	// Verify refresh token
	userID, err := m.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return m.GenerateTokenPair(userID, email, role, metadata)
}

// ExtractUserID extract user ID từ token mà không verify (dùng cho logging)
func (m *Manager) ExtractUserID(tokenString string) string {
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.SecretKey), nil
	})

	if claims, ok := token.Claims.(*Claims); ok {
		return claims.UserID
	}

	return ""
}

// GetTokenExpiry lấy thời gian hết hạn của token
func (m *Manager) GetTokenExpiry(tokenString string) (time.Time, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		return time.Time{}, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims.ExpiresAt.Time, nil
	}

	return time.Time{}, ErrInvalidToken
}

// IsTokenExpired kiểm tra token đã hết hạn chưa
func (m *Manager) IsTokenExpired(tokenString string) bool {
	expiry, err := m.GetTokenExpiry(tokenString)
	if err != nil {
		return true
	}
	return time.Now().After(expiry)
}
