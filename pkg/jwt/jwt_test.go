package jwt_test

import (
	"testing"
	"time"

	"api-core/pkg/jwt"
)

func TestGenerateAndVerifyToken(t *testing.T) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-characters-long",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test",
	})

	// Generate token
	token, err := manager.GenerateToken("user123", "user@test.com", "admin", nil)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Expected token to be non-empty")
	}

	// Verify token
	claims, err := manager.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	// Check claims
	if claims.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", claims.UserID)
	}

	if claims.Email != "user@test.com" {
		t.Errorf("Expected email 'user@test.com', got '%s'", claims.Email)
	}

	if claims.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", claims.Role)
	}
}

func TestGenerateTokenPair(t *testing.T) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey: "test-secret-key-at-least-32-characters-long",
	})

	tokens, err := manager.GenerateTokenPair("user123", "user@test.com", "user", nil)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Expected access token to be non-empty")
	}

	if tokens.RefreshToken == "" {
		t.Error("Expected refresh token to be non-empty")
	}

	if tokens.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", tokens.TokenType)
	}
}

func TestVerifyExpiredToken(t *testing.T) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey:           "test-secret-key-at-least-32-characters-long",
		AccessTokenDuration: 1 * time.Millisecond, // Very short duration
	})

	// Generate token
	token, err := manager.GenerateToken("user123", "user@test.com", "admin", nil)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Verify should fail with expired error
	_, err = manager.VerifyToken(token)
	if err != jwt.ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

func TestVerifyInvalidToken(t *testing.T) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey: "test-secret-key-at-least-32-characters-long",
	})

	// Invalid token
	_, err := manager.VerifyToken("invalid.token.string")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestTokenWithMetadata(t *testing.T) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey: "test-secret-key-at-least-32-characters-long",
	})

	metadata := map[string]interface{}{
		"name":     "John Doe",
		"verified": true,
		"age":      30,
	}

	token, err := manager.GenerateToken("user123", "user@test.com", "admin", metadata)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := manager.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if claims.Metadata["name"] != "John Doe" {
		t.Error("Expected metadata to contain name")
	}

	if claims.Metadata["verified"] != true {
		t.Error("Expected metadata to contain verified=true")
	}
}

func TestRefreshAccessToken(t *testing.T) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-at-least-32-characters-long",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
	})

	// Generate initial tokens
	tokens, err := manager.GenerateTokenPair("user123", "user@test.com", "admin", nil)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Refresh access token
	newTokens, err := manager.RefreshAccessToken(
		tokens.RefreshToken,
		"user@test.com",
		"admin",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newTokens.AccessToken == "" {
		t.Error("Expected new access token")
	}

	// Verify new access token
	claims, err := manager.VerifyToken(newTokens.AccessToken)
	if err != nil {
		t.Fatalf("Failed to verify new token: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", claims.UserID)
	}
}

func TestExtractUserID(t *testing.T) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey: "test-secret-key-at-least-32-characters-long",
	})

	token, _ := manager.GenerateToken("user123", "user@test.com", "admin", nil)

	userID := manager.ExtractUserID(token)
	if userID != "user123" {
		t.Errorf("Expected user ID 'user123', got '%s'", userID)
	}
}

func TestIsTokenExpired(t *testing.T) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey:           "test-secret-key-at-least-32-characters-long",
		AccessTokenDuration: 1 * time.Second,
	})

	token, _ := manager.GenerateToken("user123", "user@test.com", "admin", nil)

	// Token should not be expired immediately
	time.Sleep(100 * time.Millisecond)
	if manager.IsTokenExpired(token) {
		t.Error("Token should not be expired after 100ms")
	}

	// Wait for token to expire
	time.Sleep(1 * time.Second)
	if !manager.IsTokenExpired(token) {
		t.Error("Token should be expired after 1 second")
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey: "test-secret-key-at-least-32-characters-long",
	})

	for i := 0; i < b.N; i++ {
		manager.GenerateToken("user123", "user@test.com", "admin", nil)
	}
}

func BenchmarkVerifyToken(b *testing.B) {
	manager := jwt.NewManager(jwt.Config{
		SecretKey: "test-secret-key-at-least-32-characters-long",
	})

	token, _ := manager.GenerateToken("user123", "user@test.com", "admin", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.VerifyToken(token)
	}
}
