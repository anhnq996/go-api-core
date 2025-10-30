package test

import (
	"testing"

	model "api-core/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example code showing how to use database setup and cleanup in your tests:
// func TestMyFeature(t *testing.T) {
//   config := SetupTestConfigWithDB(t, &model.User{})
//   defer CleanupTestConfig(t, config)
//   userRepo := repository.NewUserRepository(config.DB)
//   // Your test code here
// }

// TestDatabaseSetupAndCleanup demonstrates database setup and cleanup
func TestDatabaseSetupAndCleanup(t *testing.T) {
	// Setup test database
	config := SetupTestConfigWithDB(t, &model.User{})
	defer CleanupTestConfig(t, config)

	// Verify database is connected
	assert.NotNil(t, config.DB)

	sqlDB, err := config.DB.DB()
	require.NoError(t, err)

	// Ping database
	err = sqlDB.Ping()
	require.NoError(t, err, "Database should be connected")

	t.Log("✅ Database connected successfully")
}

// TestDatabaseReset demonstrates how to reset database
func TestDatabaseReset(t *testing.T) {
	config := SetupTestConfigWithDB(t, &model.User{})
	defer CleanupTestConfig(t, config)

	// Create some test data
	user := &model.User{
		Name:  "Test User",
		Email: "test@example.com",
	}

	err := config.DB.Create(user).Error
	require.NoError(t, err)

	// Verify data exists
	var count int64
	config.DB.Model(&model.User{}).Count(&count)
	assert.Equal(t, int64(1), count, "Should have 1 user")

	// Reset database
	ResetTestDB(t, config.DB, &model.User{})

	// Verify data is cleared
	config.DB.Model(&model.User{}).Count(&count)
	assert.Equal(t, int64(0), count, "Should have 0 users after reset")

	t.Log("✅ Database reset successfully")
}

// TestDatabaseClean demonstrates how to clean database
func TestDatabaseClean(t *testing.T) {
	config := SetupTestConfigWithDB(t, &model.User{})
	defer CleanupTestConfig(t, config)

	// Create some test data
	user1 := &model.User{Name: "User 1", Email: "user1@example.com"}
	user2 := &model.User{Name: "User 2", Email: "user2@example.com"}

	config.DB.Create(user1)
	config.DB.Create(user2)

	// Verify data exists
	var count int64
	config.DB.Model(&model.User{}).Count(&count)
	assert.Equal(t, int64(2), count, "Should have 2 users")

	// Clean database
	CleanTestDB(t, config.DB)

	// Verify data is cleared
	config.DB.Model(&model.User{}).Count(&count)
	assert.Equal(t, int64(0), count, "Should have 0 users after clean")

	t.Log("✅ Database cleaned successfully")
}
