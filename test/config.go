package test

import (
	"os"
	"testing"

	"api-core/config"
	"api-core/pkg/cache"
	"api-core/pkg/jwt"
	"api-core/pkg/storage"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConfig holds test configuration
type TestConfig struct {
	DB         *gorm.DB
	JWTManager *jwt.Manager
	Cache      cache.Cache
	Storage    *storage.StorageManager
}

// SetupTestConfig initializes test configuration
func SetupTestConfig(t *testing.T) *TestConfig {
	// Load test environment variables
	if err := godotenv.Load(".env.test"); err != nil {
		// If .env.test doesn't exist, use default test values
		t.Log("No .env.test file found, using default test configuration")
	}

	// Setup in-memory SQLite database for testing (pure Go, no CGO required)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable SQL logs in tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Setup JWT manager for testing
	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-min-32-chars-long",
		AccessTokenDuration:  24 * 60 * 60,     // 24 hours
		RefreshTokenDuration: 7 * 24 * 60 * 60, // 7 days
		Issuer:               "test",
	})

	// Setup mock cache (in-memory)
	mockCache := cache.NewMockCache()

	// Setup mock storage
	storageConfig := config.GetDefaultStorageConfig()
	storageManager, err := storage.NewStorageManager(storageConfig)
	if err != nil {
		t.Fatalf("Failed to setup storage manager: %v", err)
	}

	return &TestConfig{
		DB:         db,
		JWTManager: jwtManager,
		Cache:      mockCache,
		Storage:    storageManager,
	}
}

// CleanupTestConfig cleans up test resources
func CleanupTestConfig(t *testing.T, config *TestConfig) {
	if config.DB != nil {
		// Clean database before closing
		CleanTestDB(t, config.DB)

		sqlDB, err := config.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// CleanTestDB cleans/resets the test database
func CleanTestDB(t *testing.T, db *gorm.DB) {
	// Get all table names
	var tables []string
	if err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables).Error; err != nil {
		t.Logf("Warning: Failed to get table names: %v", err)
		return
	}

	// Disable foreign key checks temporarily
	db.Exec("PRAGMA foreign_keys = OFF")

	// Truncate all tables
	for _, table := range tables {
		if err := db.Exec("DELETE FROM " + table).Error; err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	db.Exec("PRAGMA foreign_keys = ON")
}

// ResetTestDB resets the database by dropping all tables and re-migrating
func ResetTestDB(t *testing.T, db *gorm.DB, models ...interface{}) {
	// Disable foreign key checks
	db.Exec("PRAGMA foreign_keys = OFF")

	// Get all table names
	var tables []string
	if err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables).Error; err == nil {
		// Drop all tables
		for _, table := range tables {
			db.Exec("DROP TABLE IF EXISTS " + table)
		}
	}

	// Re-enable foreign key checks
	db.Exec("PRAGMA foreign_keys = ON")

	// Re-migrate if models are provided
	if len(models) > 0 {
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatalf("Failed to migrate test database: %v", err)
		}
	}
}

// SetupTestDB sets up test database with migrations
func SetupTestDB(t *testing.T, db *gorm.DB, models ...interface{}) error {
	if len(models) == 0 {
		return nil
	}

	// Auto migrate all provided models
	if err := db.AutoMigrate(models...); err != nil {
		return err
	}

	return nil
}

// SetupTestConfigWithDB initializes test configuration with database migrations
func SetupTestConfigWithDB(t *testing.T, models ...interface{}) *TestConfig {
	config := SetupTestConfig(t)

	// Setup database with migrations
	if len(models) > 0 {
		if err := SetupTestDB(t, config.DB, models...); err != nil {
			t.Fatalf("Failed to setup test database: %v", err)
		}
	}

	return config
}

// GetTestEnv returns test environment variable or default value
func GetTestEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsTestMode checks if running in test mode
func IsTestMode() bool {
	return os.Getenv("GO_ENV") == "test" || os.Getenv("TEST_MODE") == "true"
}
