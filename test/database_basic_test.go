package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicDatabaseConnection tests basic database connection without migrations
func TestBasicDatabaseConnection(t *testing.T) {
	// Setup test config WITHOUT migrations
	config := SetupTestConfig(t)
	defer CleanupTestConfig(t, config)

	// Verify database is connected
	assert.NotNil(t, config.DB)

	sqlDB, err := config.DB.DB()
	require.NoError(t, err)

	// Ping database
	err = sqlDB.Ping()
	require.NoError(t, err, "Database should be connected")

	t.Log("✅ Database connected successfully (no CGO required!)")
}

// TestDatabaseCleanWithoutModels tests database clean without models
func TestDatabaseCleanWithoutModels(t *testing.T) {
	config := SetupTestConfig(t)
	defer CleanupTestConfig(t, config)

	// Clean database (should work even with no tables)
	CleanTestDB(t, config.DB)

	t.Log("✅ Database clean function works")
}
