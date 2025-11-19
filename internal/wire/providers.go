package wire

import (
	"os"
	"time"

	"api-core/config"
	"api-core/pkg/cache"
	"api-core/pkg/fcm"
	"api-core/pkg/jwt"
	"api-core/pkg/storage"
	"api-core/pkg/utils"
)

// ProvideJWTManager provides JWT manager
func ProvideJWTManager() *jwt.Manager {
	// Ưu tiên dùng RSA keys nếu có; fallback sang HMAC nếu thiếu
	privatePath := getEnv("JWT_PRIVATE_KEY_PATH", "keys/private.pem")
	publicPath := getEnv("JWT_PUBLIC_KEY_PATH", "keys/public.pem")

	return jwt.NewManager(jwt.Config{
		SecretKey:            getEnv("JWT_SECRET_KEY", ""),
		PrivateKeyPath:       privatePath,
		PublicKeyPath:        publicPath,
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

// ProvideFCMClient provides FCM client (optional, returns nil if not configured)
func ProvideFCMClient() (*fcm.Client, error) {
	credentialsFile := utils.GetEnv("FIREBASE_CREDENTIALS_FILE", "keys/firebase-credentials.json")
	timeoutSeconds := utils.GetEnvInt("FCM_TIMEOUT", 10)

	// Check if credentials file exists
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		// FCM is optional, return nil without error
		return nil, nil
	}

	config := &fcm.Config{
		CredentialsFile: credentialsFile,
		Timeout:         time.Duration(timeoutSeconds) * time.Second,
	}

	client, err := fcm.NewClient(config)
	if err != nil {
		// Return nil with error, but don't fail app initialization
		return nil, err
	}

	return client, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
