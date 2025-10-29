package wire

import (
	"os"
	"time"

	"api-core/config"
	"api-core/pkg/cache"
	"api-core/pkg/jwt"
	"api-core/pkg/storage"
)

// ProvideJWTManager provides JWT manager
func ProvideJWTManager() *jwt.Manager {
	return jwt.NewManager(jwt.Config{
		SecretKey:            getEnv("JWT_SECRET_KEY", "default-secret-key-change-this-in-production-min-32-chars"),
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "apicore",
	})
}

// ProvideJWTBlacklist provides JWT blacklist
func ProvideJWTBlacklist(cacheClient cache.Cache) *jwt.Blacklist {
	return jwt.NewBlacklist(cacheClient)
}

// ProvideStorageManager provides storage manager
func ProvideStorageManager() (*storage.StorageManager, error) {
	cfg := config.GetDefaultStorageConfig()
	return storage.NewStorageManager(cfg)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
