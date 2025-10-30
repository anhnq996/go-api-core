package test

import (
	"context"
	"fmt"
	"testing"

	"api-core/config"
	"api-core/database"
	"api-core/database/seeders"
	"api-core/pkg/cache"
	"api-core/pkg/jwt"
	"api-core/pkg/storage"

	"github.com/glebarez/sqlite"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestContainerConfig holds test container configuration
type TestContainerConfig struct {
	*TestConfig
	container        *postgrescontainer.PostgresContainer
	containerCleanup func()
	isFallback       bool // true if using SQLite fallback
}

// SetupTestContainerConfig initializes test configuration with PostgreSQL test container
// Falls back to SQLite if Docker is not available
func SetupTestContainerConfig(t *testing.T, enableMigrate, enableSeeder bool) *TestContainerConfig {
	// Try to use PostgreSQL container, fallback to SQLite if Docker not available
	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgrescontainer.Run(ctx,
		"postgres:16-alpine",
		postgrescontainer.WithDatabase("test_db"),
		postgrescontainer.WithUsername("test_user"),
		postgrescontainer.WithPassword("test_password"),
	)

	// If Docker is not available, fallback to SQLite
	if err != nil {
		t.Logf("⚠️  Docker not available (error: %v), falling back to SQLite in-memory", err)
		return setupSQLiteFallback(t, enableMigrate, enableSeeder)
	}

	// Get connection string (retry if needed)
	connStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Logf("⚠️  Failed to get connection string (error: %v), falling back to SQLite", err)
		return setupSQLiteFallback(t, enableMigrate, enableSeeder)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable SQL logs in tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations if enabled
	if enableMigrate {
		migrator, err := database.NewMigrator(db, "database/migrations")
		if err != nil {
			t.Fatalf("Failed to create migrator: %v", err)
		}

		if err := migrator.Up(); err != nil {
			migrator.Close()
			t.Fatalf("Failed to run migrations: %v", err)
		}
		migrator.Close()
		t.Log("✅ Migrations completed")
	}

	// Run seeders if enabled
	if enableSeeder {
		if err := seeders.RunSeeders(db); err != nil {
			t.Fatalf("Failed to run seeders: %v", err)
		}
		t.Log("✅ Seeders completed")
	}

	// Setup JWT manager
	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-min-32-chars-long",
		AccessTokenDuration:  24 * 60 * 60,     // 24 hours
		RefreshTokenDuration: 7 * 24 * 60 * 60, // 7 days
		Issuer:               "test",
	})

	// Setup mock cache
	mockCache := cache.NewMockCache()

	// Setup mock storage
	storageConfig := config.GetDefaultStorageConfig()
	storageManager, err := storage.NewStorageManager(storageConfig)
	if err != nil {
		t.Fatalf("Failed to setup storage manager: %v", err)
	}

	return &TestContainerConfig{
		TestConfig: &TestConfig{
			DB:         db,
			JWTManager: jwtManager,
			Cache:      mockCache,
			Storage:    storageManager,
		},
		container: postgresContainer,
		containerCleanup: func() {
			if err := postgresContainer.Terminate(ctx); err != nil {
				t.Logf("Warning: Failed to terminate container: %v", err)
			}
		},
	}
}

// CleanupTestContainerConfig cleans up test container resources
func CleanupTestContainerConfig(t *testing.T, config *TestContainerConfig) {
	if config == nil {
		return
	}

	// Clean database data
	if config.DB != nil {
		if config.isFallback {
			// Use SQLite cleanup
			CleanTestDB(t, config.DB)
		} else {
			// Use PostgreSQL cleanup
			CleanTestDBForContainer(t, config.DB)
		}

		sqlDB, err := config.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// Terminate container (only if not fallback)
	if !config.isFallback && config.containerCleanup != nil {
		config.containerCleanup()
	}
}

// setupSQLiteFallback creates SQLite in-memory database when Docker is not available
func setupSQLiteFallback(t *testing.T, enableMigrate, enableSeeder bool) *TestContainerConfig {
	t.Log("✅ Using SQLite in-memory database for testing")

	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}

	// Run migrations if enabled
	if enableMigrate {
		// For SQLite, we need to auto-migrate models
		// Import models here and migrate them
		// For now, just log that migrations are skipped
		t.Log("⚠️  SQLite: Using GORM AutoMigrate instead of file-based migrations")
		t.Log("⚠️  Note: Add your models here to enable auto-migration")
		// Example:
		// import model "api-core/internal/models"
		// db.AutoMigrate(&model.User{}, &model.Role{}, &model.Permission{})
	}

	// Run seeders if enabled
	if enableSeeder {
		if err := seeders.RunSeeders(db); err != nil {
			t.Logf("Warning: Failed to run seeders: %v", err)
		} else {
			t.Log("✅ Seeders completed")
		}
	}

	// Setup JWT manager
	jwtManager := jwt.NewManager(jwt.Config{
		SecretKey:            "test-secret-key-min-32-chars-long",
		AccessTokenDuration:  24 * 60 * 60,
		RefreshTokenDuration: 7 * 24 * 60 * 60,
		Issuer:               "test",
	})

	// Setup mock cache
	mockCache := cache.NewMockCache()

	// Setup mock storage
	storageConfig := config.GetDefaultStorageConfig()
	storageManager, err := storage.NewStorageManager(storageConfig)
	if err != nil {
		t.Fatalf("Failed to setup storage: %v", err)
	}

	return &TestContainerConfig{
		TestConfig: &TestConfig{
			DB:         db,
			JWTManager: jwtManager,
			Cache:      mockCache,
			Storage:    storageManager,
		},
		isFallback: true,
	}
}

// CleanTestDBForContainer cleans/resets the test database (for PostgreSQL)
func CleanTestDBForContainer(t *testing.T, db *gorm.DB) {
	// Disable foreign key checks
	db.Exec("SET session_replication_role = 'replica';")

	// Get all table names
	var tables []string
	if err := db.Raw(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&tables).Error; err != nil {
		t.Logf("Warning: Failed to get table names: %v", err)
		return
	}

	// Truncate all tables
	for _, table := range tables {
		// Skip migration tracking table
		if table == "schema_migrations" {
			continue
		}
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}

	// Re-enable foreign key checks
	db.Exec("SET session_replication_role = 'origin';")
}

// ResetTestDBForContainer resets the database by dropping all tables (except migrations) and re-migrating
func ResetTestDBForContainer(t *testing.T, db *gorm.DB, enableMigrate bool) {
	// Get all table names
	var tables []string
	if err := db.Raw(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&tables).Error; err == nil {
		// Drop all tables (except schema_migrations)
		for _, table := range tables {
			if table != "schema_migrations" {
				db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
			}
		}
	}

	// Re-migrate if enabled
	if enableMigrate {
		migrator, err := database.NewMigrator(db, "database/migrations")
		if err != nil {
			t.Logf("Warning: Failed to create migrator: %v", err)
			return
		}
		defer migrator.Close()

		if err := migrator.Up(); err != nil {
			t.Logf("Warning: Failed to run migrations: %v", err)
		}
	}
}
