package interfaces

import (
	"context"
	"io"
)

// FileInfo chứa thông tin về file
type FileInfo struct {
	Name         string            `json:"name"`          // Tên file
	Size         int64             `json:"size"`          // Kích thước file (bytes)
	ContentType  string            `json:"content_type"`  // MIME type
	Path         string            `json:"path"`          // Đường dẫn file
	URL          string            `json:"url"`           // URL để truy cập file
	ETag         string            `json:"etag"`          // ETag cho cache
	LastModified int64             `json:"last_modified"` // Timestamp last modified
	Metadata     map[string]string `json:"metadata"`      // Metadata tùy chỉnh
}

// UploadOptions tùy chọn khi upload file
type UploadOptions struct {
	Path         string            `json:"path"`          // Đường dẫn lưu trữ
	ContentType  string            `json:"content_type"`  // MIME type
	Public       bool              `json:"public"`        // File có public không
	Metadata     map[string]string `json:"metadata"`      // Metadata tùy chỉnh
	ACL          string            `json:"acl"`           // Access Control List (cho S3)
	CacheControl string            `json:"cache_control"` // Cache control header
}

// ListOptions tùy chọn khi list files
type ListOptions struct {
	Prefix    string `json:"prefix"`    // Prefix để filter
	Delimiter string `json:"delimiter"` // Delimiter cho folder structure
	MaxKeys   int    `json:"max_keys"`  // Số lượng file tối đa
	Marker    string `json:"marker"`    // Marker cho pagination
}

// ListResult kết quả khi list files
type ListResult struct {
	Files       []FileInfo `json:"files"`        // Danh sách files
	Folders     []string   `json:"folders"`      // Danh sách folders
	NextMarker  string     `json:"next_marker"`  // Marker cho trang tiếp theo
	IsTruncated bool       `json:"is_truncated"` // Còn file nữa không
}

// Storage interface định nghĩa các method cần thiết cho storage
type Storage interface {
	// Upload file từ io.Reader
	Upload(ctx context.Context, key string, reader io.Reader, options *UploadOptions) (*FileInfo, error)

	// Upload file từ bytes
	UploadBytes(ctx context.Context, key string, data []byte, options *UploadOptions) (*FileInfo, error)

	// Download file về io.Reader
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Download file về bytes
	DownloadBytes(ctx context.Context, key string) ([]byte, error)

	// Get file info
	GetInfo(ctx context.Context, key string) (*FileInfo, error)

	// Check file exists
	Exists(ctx context.Context, key string) (bool, error)

	// Delete file
	Delete(ctx context.Context, key string) error

	// Delete multiple files
	DeleteMultiple(ctx context.Context, keys []string) error

	// List files
	List(ctx context.Context, options *ListOptions) (*ListResult, error)

	// Get public URL
	GetURL(ctx context.Context, key string) (string, error)

	// Get signed URL (cho private files)
	GetSignedURL(ctx context.Context, key string, expiresIn int64) (string, error)

	// Copy file
	Copy(ctx context.Context, srcKey, dstKey string) error

	// Move file (copy + delete)
	Move(ctx context.Context, srcKey, dstKey string) error
}

// ImageProcessor interface cho xử lý ảnh
type ImageProcessor interface {
	// Resize ảnh
	Resize(ctx context.Context, reader io.Reader, width, height int) (io.Reader, error)

	// Crop ảnh
	Crop(ctx context.Context, reader io.Reader, x, y, width, height int) (io.Reader, error)

	// Thêm watermark
	AddWatermark(ctx context.Context, reader io.Reader, watermarkPath string, position string) (io.Reader, error)

	// Convert format
	Convert(ctx context.Context, reader io.Reader, format string) (io.Reader, error)

	// Get image info
	GetInfo(ctx context.Context, reader io.Reader) (*ImageInfo, error)
}

// ImageInfo thông tin về ảnh
type ImageInfo struct {
	Width      int    `json:"width"`       // Chiều rộng
	Height     int    `json:"height"`      // Chiều cao
	Format     string `json:"format"`      // Định dạng (JPEG, PNG, etc.)
	ColorModel string `json:"color_model"` // Color model (RGB, RGBA, etc.)
	Size       int64  `json:"size"`        // Kích thước file
}

// FileValidator interface cho validation file
type FileValidator interface {
	// Validate file type
	ValidateType(contentType string, allowedTypes []string) error

	// Validate file size
	ValidateSize(size int64, maxSize int64) error

	// Validate file content (check magic bytes)
	ValidateContent(reader io.Reader, expectedType string) error

	// Validate file hoàn chỉnh
	ValidateFile(filename, contentType string, size int64, reader io.Reader, category string) error

	// Check if file is image
	IsImage(contentType string) bool
}
