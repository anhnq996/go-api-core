package local

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"anhnq/api-core/pkg/storage/interfaces"
)

// LocalStorage implementation cho local file system
type LocalStorage struct {
	basePath string // Đường dẫn gốc để lưu files
	baseURL  string // Base URL để truy cập files
}

// NewLocalStorage tạo instance mới của LocalStorage
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	// Tạo thư mục base nếu chưa tồn tại
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// Upload file từ io.Reader
func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, options *interfaces.UploadOptions) (*interfaces.FileInfo, error) {
	// Tạo đường dẫn đầy đủ
	fullPath := filepath.Join(s.basePath, key)

	// Tạo thư mục nếu chưa tồn tại
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Tạo file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data từ reader vào file
	size, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(fullPath) // Cleanup nếu có lỗi
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Determine content type
	contentType := options.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(key))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	// Generate URL
	fileURL := s.generateURL(key)

	return &interfaces.FileInfo{
		Name:         filepath.Base(key),
		Size:         size,
		ContentType:  contentType,
		Path:         key,
		URL:          fileURL,
		ETag:         s.generateETag(fullPath),
		LastModified: info.ModTime().Unix(),
		Metadata:     options.Metadata,
	}, nil
}

// UploadBytes upload file từ bytes
func (s *LocalStorage) UploadBytes(ctx context.Context, key string, data []byte, options *interfaces.UploadOptions) (*interfaces.FileInfo, error) {
	return s.Upload(ctx, key, strings.NewReader(string(data)), options)
}

// Download file về io.Reader
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// DownloadBytes download file về bytes
func (s *LocalStorage) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	reader, err := s.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// GetInfo lấy thông tin file
func (s *LocalStorage) GetInfo(ctx context.Context, key string) (*interfaces.FileInfo, error) {
	fullPath := filepath.Join(s.basePath, key)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	contentType := mime.TypeByExtension(filepath.Ext(key))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &interfaces.FileInfo{
		Name:         filepath.Base(key),
		Size:         info.Size(),
		ContentType:  contentType,
		Path:         key,
		URL:          s.generateURL(key),
		ETag:         s.generateETag(fullPath),
		LastModified: info.ModTime().Unix(),
		Metadata:     make(map[string]string),
	}, nil
}

// Exists kiểm tra file có tồn tại không
func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(s.basePath, key)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return true, nil
}

// Delete xóa file
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, key)

	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", key)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteMultiple xóa nhiều files
func (s *LocalStorage) DeleteMultiple(ctx context.Context, keys []string) error {
	var errors []string

	for _, key := range keys {
		if err := s.Delete(ctx, key); err != nil {
			errors = append(errors, fmt.Sprintf("failed to delete %s: %v", key, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete some files: %s", strings.Join(errors, "; "))
	}

	return nil
}

// List list files
func (s *LocalStorage) List(ctx context.Context, options *interfaces.ListOptions) (*interfaces.ListResult, error) {
	searchPath := s.basePath
	if options.Prefix != "" {
		searchPath = filepath.Join(s.basePath, options.Prefix)
	}

	var files []interfaces.FileInfo
	var folders []string
	var nextMarker string
	isTruncated := false

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip base directory
		if path == searchPath {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(s.basePath, path)
		if err != nil {
			return err
		}

		// Apply prefix filter
		if options.Prefix != "" && !strings.HasPrefix(relPath, options.Prefix) {
			return nil
		}

		// Check max keys limit
		if options.MaxKeys > 0 && len(files) >= options.MaxKeys {
			isTruncated = true
			nextMarker = relPath
			return filepath.SkipDir
		}

		if info.IsDir() {
			// Handle delimiter for folder structure
			if options.Delimiter != "" {
				// Check if this is a direct child of the search path
				relDir := filepath.Dir(relPath)
				if relDir == options.Prefix || (options.Prefix == "" && relDir == ".") {
					folders = append(folders, relPath)
					return filepath.SkipDir // Don't walk into subdirectories
				}
			}
		} else {
			// It's a file
			contentType := mime.TypeByExtension(filepath.Ext(path))
			if contentType == "" {
				contentType = "application/octet-stream"
			}

			files = append(files, interfaces.FileInfo{
				Name:         info.Name(),
				Size:         info.Size(),
				ContentType:  contentType,
				Path:         relPath,
				URL:          s.generateURL(relPath),
				ETag:         s.generateETag(path),
				LastModified: info.ModTime().Unix(),
				Metadata:     make(map[string]string),
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return &interfaces.ListResult{
		Files:       files,
		Folders:     folders,
		NextMarker:  nextMarker,
		IsTruncated: isTruncated,
	}, nil
}

// GetURL lấy public URL
func (s *LocalStorage) GetURL(ctx context.Context, key string) (string, error) {
	return s.generateURL(key), nil
}

// GetSignedURL lấy signed URL (local storage không cần signed URL)
func (s *LocalStorage) GetSignedURL(ctx context.Context, key string, expiresIn int64) (string, error) {
	return s.generateURL(key), nil
}

// Copy copy file
func (s *LocalStorage) Copy(ctx context.Context, srcKey, dstKey string) error {
	srcPath := filepath.Join(s.basePath, srcKey)
	dstPath := filepath.Join(s.basePath, dstKey)

	// Tạo thư mục đích nếu chưa tồn tại
	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Copy file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Move move file (copy + delete)
func (s *LocalStorage) Move(ctx context.Context, srcKey, dstKey string) error {
	if err := s.Copy(ctx, srcKey, dstKey); err != nil {
		return err
	}

	return s.Delete(ctx, srcKey)
}

// generateURL tạo URL cho file
func (s *LocalStorage) generateURL(key string) string {
	if s.baseURL == "" {
		return "/" + key
	}

	// Ensure baseURL ends with /
	if !strings.HasSuffix(s.baseURL, "/") {
		s.baseURL += "/"
	}

	return s.baseURL + key
}

// generateETag tạo ETag cho file
func (s *LocalStorage) generateETag(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().Unix())
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Sprintf("%d", time.Now().Unix())
	}

	return fmt.Sprintf("\"%x\"", hash.Sum(nil))
}
