package test

import (
	"testing"

	model "api-core/internal/models"
	repository "api-core/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContainerWithMigrateAndSeeder demonstrates full test container setup
func TestContainerWithMigrateAndSeeder(t *testing.T) {
	// Setup test container with migrations and seeders
	config := SetupTestContainerConfig(t, true, true) // enableMigrate=true, enableSeeder=true
	defer CleanupTestContainerConfig(t, config)

	// Verify database is connected
	assert.NotNil(t, config.DB)

	// Skip PostgreSQL-specific checks if using fallback
	if config.isFallback {
		t.Log("✅ Using SQLite fallback (Docker not available)")
		return
	}

	// Verify migrations ran (check if tables exist) - PostgreSQL only
	var count int64
	err := config.DB.Table("information_schema.tables").
		Where("table_schema = ? AND table_name = ?", "public", "users").
		Count(&count).Error
	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Users table should exist")

	// Verify seeders ran (check if roles exist)
	var roleCount int64
	err = config.DB.Model(&model.Role{}).Count(&roleCount).Error
	require.NoError(t, err)
	assert.Greater(t, roleCount, int64(0), "Roles should be seeded")

	t.Log("✅ Test container with migrations and seeders setup successfully")
}

// TestContainerWithMigrateOnly demonstrates test container with migrations only
func TestContainerWithMigrateOnly(t *testing.T) {
	// Setup test container with migrations only (no seeders)
	config := SetupTestContainerConfig(t, true, false) // enableMigrate=true, enableSeeder=false
	defer CleanupTestContainerConfig(t, config)

	// Verify database is connected
	assert.NotNil(t, config.DB)

	// Verify migrations ran
	var count int64
	err := config.DB.Table("information_schema.tables").
		Where("table_schema = ? AND table_name = ?", "public", "users").
		Count(&count).Error
	require.NoError(t, err)
	assert.Greater(t, count, int64(0), "Users table should exist")

	// Verify seeders did NOT run (roles should be empty)
	var roleCount int64
	err = config.DB.Model(&model.Role{}).Count(&roleCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), roleCount, "Roles should NOT be seeded")

	t.Log("✅ Test container with migrations only setup successfully")
}

// TestContainerDatabaseCleanup demonstrates database cleanup after test
func TestContainerDatabaseCleanup(t *testing.T) {
	// Setup test container with migrations and seeders
	config := SetupTestContainerConfig(t, false, false) // Skip migrate/seeder for this test
	defer CleanupTestContainerConfig(t, config)

	// Skip if fallback (SQLite cleanup tested separately)
	if config.isFallback {
		t.Log("✅ Using SQLite fallback, cleanup tested separately")
		return
	}

	// Create test user
	repository.NewUserRepository(config.DB)
	user := &model.User{
		Name:  "Test User",
		Email: "test@example.com",
	}
	err := config.DB.Create(user).Error
	require.NoError(t, err)

	// Verify user exists
	var userCount int64
	config.DB.Model(&model.User{}).Count(&userCount)
	assert.Equal(t, int64(1), userCount, "Should have 1 user")

	// Clean database
	CleanTestDBForContainer(t, config.DB)

	// Verify data is cleaned (but tables still exist)
	config.DB.Model(&model.User{}).Count(&userCount)
	assert.Equal(t, int64(0), userCount, "Should have 0 users after clean")

	// Verify tables still exist
	var tableCount int64
	config.DB.Table("information_schema.tables").
		Where("table_schema = ? AND table_name = ?", "public", "users").
		Count(&tableCount)
	assert.Greater(t, tableCount, int64(0), "Table should still exist")

	t.Log("✅ Database cleanup works correctly")
}

// Example code showing how to use container in tests:
// func TestMyFeature(t *testing.T) {
//   config := SetupTestContainerConfig(t, true, true) // Auto migrate + seed
//   defer CleanupTestContainerConfig(t, config)        // Auto cleanup
//
//   userRepo := repository.NewUserRepository(config.DB)
//   users, err := userRepo.FindAll(nil)
//   // Your test code here
// }
