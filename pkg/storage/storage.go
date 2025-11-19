package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"api-core/config"
	"api-core/pkg/storage/aws"
	"api-core/pkg/storage/image"
	"api-core/pkg/storage/interfaces"
	"api-core/pkg/storage/local"
	"api-core/pkg/storage/validator"

	"github.com/google/uuid"
)

// StorageManager quản lý storage operations
type StorageManager struct {
	storage        interfaces.Storage
	imageProcessor interfaces.ImageProcessor
	validator      interfaces.FileValidator
}

// UploadResult kết quả upload file
type UploadResult struct {
	Path        string `json:"path"`         // Đường dẫn file
	URL         string `json:"url"`          // URL để truy cập file
	Size        int64  `json:"size"`         // Kích thước file
	ContentType string `json:"content_type"` // MIME type
	ETag        string `json:"etag"`         // ETag
}

// UploadOptions tùy chọn upload
type UploadOptions struct {
	Category     string            `json:"category"`      // image, document, video, audio, archive
	Path         string            `json:"path"`          // Custom path
	Public       bool              `json:"public"`        // Public access
	ProcessImage bool              `json:"process_image"` // Process image (resize, etc.)
	ImageOptions *ImageOptions     `json:"image_options"` // Image processing options
	Metadata     map[string]string `json:"metadata"`      // Custom metadata
}

// ImageOptions tùy chọn xử lý ảnh
type ImageOptions struct {
	Resize    *ResizeOptions    `json:"resize"`    // Resize options
	Crop      *CropOptions      `json:"crop"`      // Crop options
	Watermark *WatermarkOptions `json:"watermark"` // Watermark options
	Format    string            `json:"format"`    // Output format
	Quality   int               `json:"quality"`   // JPEG quality
}

// ResizeOptions tùy chọn resize
type ResizeOptions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// CropOptions tùy chọn crop
type CropOptions struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// WatermarkOptions tùy chọn watermark
type WatermarkOptions struct {
	Path     string `json:"path"`     // Path to watermark image
	Position string `json:"position"` // top-left, top-right, bottom-left, bottom-right, center
}

// NewStorageManager tạo instance mới của StorageManager
func NewStorageManager(cfg config.StorageConfig) (*StorageManager, error) {
	// Tạo storage
	var storage interfaces.Storage
	var err error

	switch cfg.Driver {
	case "local":
		storage, err = local.NewLocalStorage(cfg.Local.BasePath, cfg.Local.BaseURL)
	case "s3":
		storage, err = aws.NewS3Storage(aws.S3Config{
			Bucket:          cfg.S3.Bucket,
			Region:          cfg.S3.Region,
			AccessKeyID:     cfg.S3.AccessKeyID,
			SecretAccessKey: cfg.S3.SecretAccessKey,
			BaseURL:         cfg.S3.BaseURL,
		})
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	// Tạo image processor
	imageProcessor := image.NewImageProcessor(cfg.Image.Quality)

	// Tạo file validator
	fileValidator := validator.NewFileValidator()
	fileValidator.SetMaxSize("default", cfg.Validation.MaxFileSize)

	return &StorageManager{
		storage:        storage,
		imageProcessor: imageProcessor,
		validator:      fileValidator,
	}, nil
}

// UploadFile upload file từ multipart form
func (sm *StorageManager) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, options *UploadOptions) (*UploadResult, error) {
	// Mở file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Đọc file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Validate file
	if err := sm.validator.ValidateFile(
		fileHeader.Filename,
		fileHeader.Header.Get("Content-Type"),
		fileHeader.Size,
		strings.NewReader(string(content)),
		options.Category,
	); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Generate unique filename
	filename := sm.generateFilename(fileHeader.Filename)
	path := sm.generatePath(options.Path, filename)

	// Process image if needed
	var processedContent []byte
	if options.ProcessImage && sm.validator.IsImage(fileHeader.Header.Get("Content-Type")) {
		processedContent, err = sm.processImage(content, options.ImageOptions)
		if err != nil {
			return nil, fmt.Errorf("image processing failed: %w", err)
		}
	} else {
		processedContent = content
	}

	// Prepare upload options
	uploadOptions := &interfaces.UploadOptions{
		Path:        path,
		ContentType: fileHeader.Header.Get("Content-Type"),
		Public:      options.Public,
		Metadata:    options.Metadata,
	}

	// Upload to storage
	fileInfo, err := sm.storage.UploadBytes(ctx, path, processedContent, uploadOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadResult{
		Path:        fileInfo.Path,
		URL:         fileInfo.URL,
		Size:        fileInfo.Size,
		ContentType: fileInfo.ContentType,
		ETag:        fileInfo.ETag,
	}, nil
}

// UploadBytes upload file từ bytes
func (sm *StorageManager) UploadBytes(ctx context.Context, filename string, content []byte, contentType string, options *UploadOptions) (*UploadResult, error) {
	// Validate file
	if err := sm.validator.ValidateFile(
		filename,
		contentType,
		int64(len(content)),
		strings.NewReader(string(content)),
		options.Category,
	); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Generate unique filename
	uniqueFilename := sm.generateFilename(filename)
	path := sm.generatePath(options.Path, uniqueFilename)

	// Process image if needed
	var processedContent []byte
	var err error
	if options.ProcessImage && sm.validator.IsImage(contentType) {
		processedContent, err = sm.processImage(content, options.ImageOptions)
		if err != nil {
			return nil, fmt.Errorf("image processing failed: %w", err)
		}
	} else {
		processedContent = content
	}

	// Prepare upload options
	uploadOptions := &interfaces.UploadOptions{
		Path:        path,
		ContentType: contentType,
		Public:      options.Public,
		Metadata:    options.Metadata,
	}

	// Upload to storage
	fileInfo, err := sm.storage.UploadBytes(ctx, path, processedContent, uploadOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &UploadResult{
		Path:        fileInfo.Path,
		URL:         fileInfo.URL,
		Size:        fileInfo.Size,
		ContentType: fileInfo.ContentType,
		ETag:        fileInfo.ETag,
	}, nil
}

// DeleteFile xóa file
func (sm *StorageManager) DeleteFile(ctx context.Context, path string) error {
	return sm.storage.Delete(ctx, path)
}

// GetFileURL lấy URL của file
func (sm *StorageManager) GetFileURL(ctx context.Context, path string, signed bool, expiresIn int64) (string, error) {
	if signed {
		return sm.storage.GetSignedURL(ctx, path, expiresIn)
	}
	return sm.storage.GetURL(ctx, path)
}

// CopyFile copy file
func (sm *StorageManager) CopyFile(ctx context.Context, srcPath, dstPath string) error {
	return sm.storage.Copy(ctx, srcPath, dstPath)
}

// MoveFile move file (copy + delete)
func (sm *StorageManager) MoveFile(ctx context.Context, srcPath, dstPath string) error {
	return sm.storage.Move(ctx, srcPath, dstPath)
}

// FileExists kiểm tra file có tồn tại không
func (sm *StorageManager) FileExists(ctx context.Context, path string) (bool, error) {
	return sm.storage.Exists(ctx, path)
}

// GetFileInfo lấy thông tin file
func (sm *StorageManager) GetFileInfo(ctx context.Context, path string) (*interfaces.FileInfo, error) {
	return sm.storage.GetInfo(ctx, path)
}

// processImage xử lý ảnh
func (sm *StorageManager) processImage(content []byte, options *ImageOptions) ([]byte, error) {
	if sm.imageProcessor == nil {
		return content, nil
	}

	var reader io.Reader = strings.NewReader(string(content))

	// Apply image processing
	if options != nil {
		// Resize
		if options.Resize != nil {
			var err error
			reader, err = sm.imageProcessor.Resize(context.Background(), reader, options.Resize.Width, options.Resize.Height)
			if err != nil {
				return nil, fmt.Errorf("failed to resize image: %w", err)
			}
		}

		// Crop
		if options.Crop != nil {
			var err error
			reader, err = sm.imageProcessor.Crop(context.Background(), reader, options.Crop.X, options.Crop.Y, options.Crop.Width, options.Crop.Height)
			if err != nil {
				return nil, fmt.Errorf("failed to crop image: %w", err)
			}
		}

		// Watermark
		if options.Watermark != nil {
			var err error
			reader, err = sm.imageProcessor.AddWatermark(context.Background(), reader, options.Watermark.Path, options.Watermark.Position)
			if err != nil {
				return nil, fmt.Errorf("failed to add watermark: %w", err)
			}
		}

		// Convert format
		if options.Format != "" {
			var err error
			reader, err = sm.imageProcessor.Convert(context.Background(), reader, options.Format)
			if err != nil {
				return nil, fmt.Errorf("failed to convert image: %w", err)
			}
		}
	}

	// Read processed content
	processedContent, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read processed image: %w", err)
	}

	return processedContent, nil
}

// generateFilename tạo filename unique
func (sm *StorageManager) generateFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := strings.TrimSuffix(strings.ReplaceAll(originalFilename, " ", "-"), ext)

	// Generate unique name
	uniqueID := uuid.New().String()
	return fmt.Sprintf("%s_%s%s", name, uniqueID, ext)
}

// generatePath tạo path cho file
func (sm *StorageManager) generatePath(customPath, filename string) string {
	if customPath != "" {
		if !strings.HasSuffix(customPath, "/") {
			customPath += "/"
		}
		return customPath + filename
	}

	// Default path structure: year/month/filename
	now := time.Now()
	return fmt.Sprintf("%d/%02d/%s", now.Year(), now.Month(), filename)
}

// GetDefaultUploadOptions tạo upload options mặc định
func GetDefaultUploadOptions(category string) *UploadOptions {
	return &UploadOptions{
		Category:     category,
		Public:       true,
		ProcessImage: false,
		Metadata:     make(map[string]string),
	}
}

// GetImageUploadOptions tạo upload options cho ảnh
func GetImageUploadOptions(width, height int, quality int) *UploadOptions {
	if quality <= 0 || quality > 100 {
		quality = 90
	}

	return &UploadOptions{
		Category:     "image",
		Public:       true,
		ProcessImage: true,
		ImageOptions: &ImageOptions{
			Resize: &ResizeOptions{
				Width:  width,
				Height: height,
			},
			Quality: quality,
		},
		Metadata: make(map[string]string),
	}
}
