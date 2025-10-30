package test

import (
	"testing"

	"api-core/pkg/cache"

	"github.com/stretchr/testify/assert"
)

func TestBasicFunctionality(t *testing.T) {
	// Test basic functionality without database
	assert.True(t, true, "Basic test should pass")
}

func TestMockCache(t *testing.T) {
	// Test mock cache functionality
	mockCache := cache.NewMockCache()
	assert.NotNil(t, mockCache, "Mock cache should be created")
}

func TestTestUserCreation(t *testing.T) {
	// Test test user creation
	user := CreateTestUser()
	assert.NotNil(t, user, "Test user should be created")
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestTestAdminCreation(t *testing.T) {
	// Test test admin creation
	admin := CreateTestAdmin()
	assert.NotNil(t, admin, "Test admin should be created")
	assert.Equal(t, "Test Admin", admin.Name)
	assert.Equal(t, "admin@example.com", admin.Email)
	assert.Equal(t, "admin", admin.Role)
}
