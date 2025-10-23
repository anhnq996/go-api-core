package storage

import (
	"fmt"

	"anhnq/api-core/config"
	"anhnq/api-core/pkg/storage/aws"
	"anhnq/api-core/pkg/storage/image"
	"anhnq/api-core/pkg/storage/interfaces"
	"anhnq/api-core/pkg/storage/local"
	"anhnq/api-core/pkg/storage/validator"
)

// StorageFactory factory để tạo storage instances
type StorageFactory struct{}

// NewStorageFactory tạo instance mới của StorageFactory
func NewStorageFactory() *StorageFactory {
	return &StorageFactory{}
}

// CreateStorage tạo storage instance dựa trên config
func (f *StorageFactory) CreateStorage(cfg config.StorageConfig) (interfaces.Storage, error) {
	switch cfg.Driver {
	case "local":
		return local.NewLocalStorage(cfg.Local.BasePath, cfg.Local.BaseURL)
	case "s3":
		return aws.NewS3Storage(aws.S3Config{
			Bucket:          cfg.S3.Bucket,
			Region:          cfg.S3.Region,
			AccessKeyID:     cfg.S3.AccessKeyID,
			SecretAccessKey: cfg.S3.SecretAccessKey,
			BaseURL:         cfg.S3.BaseURL,
		})
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}

// CreateImageProcessor tạo image processor
func (f *StorageFactory) CreateImageProcessor(cfg config.StorageConfig) interfaces.ImageProcessor {
	return image.NewImageProcessor(cfg.Image.Quality)
}

// CreateFileValidator tạo file validator
func (f *StorageFactory) CreateFileValidator(cfg config.StorageConfig) interfaces.FileValidator {
	validator := validator.NewFileValidator()

	// Set max file size
	validator.SetMaxSize("default", cfg.Validation.MaxFileSize)

	return validator
}

// CreateStorageComponents tạo tất cả storage components
func (f *StorageFactory) CreateStorageComponents(cfg config.StorageConfig) (
	interfaces.Storage,
	interfaces.ImageProcessor,
	interfaces.FileValidator,
	error,
) {
	// Create storage
	storage, err := f.CreateStorage(cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Create image processor
	imageProcessor := f.CreateImageProcessor(cfg)

	// Create file validator
	fileValidator := f.CreateFileValidator(cfg)

	return storage, imageProcessor, fileValidator, nil
}
