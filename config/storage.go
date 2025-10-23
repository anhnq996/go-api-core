package config

import (
	"fmt"
	"os"
	"strconv"
)

// StorageConfig cấu hình cho storage
type StorageConfig struct {
	Driver     string           `json:"driver"` // local, s3
	Local      LocalConfig      `json:"local"`
	S3         S3Config         `json:"s3"`
	Image      ImageConfig      `json:"image"`
	Validation ValidationConfig `json:"validation"`
}

// LocalConfig cấu hình cho local storage
type LocalConfig struct {
	BasePath string `json:"base_path"`
	BaseURL  string `json:"base_url"`
}

// S3Config cấu hình cho S3 storage
type S3Config struct {
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BaseURL         string `json:"base_url"`
}

// ImageConfig cấu hình cho image processing
type ImageConfig struct {
	Quality int `json:"quality"`
}

// ValidationConfig cấu hình cho file validation
type ValidationConfig struct {
	MaxFileSize int64 `json:"max_file_size"`
}

// GetDefaultStorageConfig lấy cấu hình storage mặc định
func GetDefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Driver: getEnvStorage("STORAGE_DRIVER", "local"),
		Local: LocalConfig{
			BasePath: getEnvStorage("STORAGE_LOCAL_PATH", "storages/app"),
			BaseURL:  getEnvStorage("STORAGE_LOCAL_URL", "/storages"),
		},
		S3: S3Config{
			Bucket:          getEnvStorage("STORAGE_S3_BUCKET", ""),
			Region:          getEnvStorage("STORAGE_S3_REGION", "us-east-1"),
			AccessKeyID:     getEnvStorage("STORAGE_S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnvStorage("STORAGE_S3_SECRET_ACCESS_KEY", ""),
			BaseURL:         getEnvStorage("STORAGE_S3_BASE_URL", ""),
		},
		Image: ImageConfig{
			Quality: getEnvIntStorage("STORAGE_IMAGE_QUALITY", 90),
		},
		Validation: ValidationConfig{
			MaxFileSize: getEnvInt64Storage("STORAGE_MAX_FILE_SIZE", 10*1024*1024), // 10MB
		},
	}
}

// getEnvStorage lấy environment variable với default value
func getEnvStorage(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntStorage lấy environment variable dạng int với default value
func getEnvIntStorage(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvInt64Storage lấy environment variable dạng int64 với default value
func getEnvInt64Storage(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// ValidateStorageConfig validate storage config
func ValidateStorageConfig(config StorageConfig) error {
	switch config.Driver {
	case "local":
		if config.Local.BasePath == "" {
			return fmt.Errorf("local storage base path is required")
		}
	case "s3":
		if config.S3.Bucket == "" {
			return fmt.Errorf("S3 bucket is required")
		}
		if config.S3.Region == "" {
			return fmt.Errorf("S3 region is required")
		}
		if config.S3.AccessKeyID == "" {
			return fmt.Errorf("S3 access key ID is required")
		}
		if config.S3.SecretAccessKey == "" {
			return fmt.Errorf("S3 secret access key is required")
		}
	default:
		return fmt.Errorf("unsupported storage driver: %s", config.Driver)
	}

	if config.Image.Quality < 1 || config.Image.Quality > 100 {
		return fmt.Errorf("image quality must be between 1 and 100")
	}

	if config.Validation.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be greater than 0")
	}

	return nil
}
