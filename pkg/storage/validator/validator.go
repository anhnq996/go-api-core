package validator

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
)

// FileValidator implementation cho validation file
type FileValidator struct {
	allowedTypes map[string][]string // map[category][]mime_types
	maxSizes     map[string]int64    // map[category]max_size_in_bytes
}

// NewFileValidator tạo instance mới của FileValidator
func NewFileValidator() *FileValidator {
	return &FileValidator{
		allowedTypes: map[string][]string{
			"image": {
				"image/jpeg",
				"image/jpg",
				"image/png",
				"image/gif",
				"image/webp",
				"image/bmp",
				"image/tiff",
			},
			"document": {
				"application/pdf",
				"application/msword",
				"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				"application/vnd.ms-excel",
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
				"application/vnd.ms-powerpoint",
				"application/vnd.openxmlformats-officedocument.presentationml.presentation",
				"text/plain",
				"text/csv",
			},
			"video": {
				"video/mp4",
				"video/avi",
				"video/mov",
				"video/wmv",
				"video/flv",
				"video/webm",
				"video/mkv",
			},
			"audio": {
				"audio/mp3",
				"audio/mpeg",
				"audio/wav",
				"audio/ogg",
				"audio/aac",
				"audio/flac",
			},
			"archive": {
				"application/zip",
				"application/x-rar-compressed",
				"application/x-7z-compressed",
				"application/gzip",
				"application/x-tar",
			},
		},
		maxSizes: map[string]int64{
			"image":    10 * 1024 * 1024,  // 10MB
			"document": 50 * 1024 * 1024,  // 50MB
			"video":    500 * 1024 * 1024, // 500MB
			"audio":    100 * 1024 * 1024, // 100MB
			"archive":  200 * 1024 * 1024, // 200MB
			"default":  10 * 1024 * 1024,  // 10MB
		},
	}
}

// ValidateType kiểm tra file type
func (v *FileValidator) ValidateType(contentType string, allowedTypes []string) error {
	if len(allowedTypes) == 0 {
		return fmt.Errorf("no allowed types specified")
	}

	// Check if content type is in allowed list
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("file type %s is not allowed. Allowed types: %s", contentType, strings.Join(allowedTypes, ", "))
}

// ValidateSize kiểm tra file size
func (v *FileValidator) ValidateSize(size int64, maxSize int64) error {
	if maxSize <= 0 {
		return fmt.Errorf("invalid max size: %d", maxSize)
	}

	if size > maxSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size %d bytes", size, maxSize)
	}

	return nil
}

// ValidateContent kiểm tra file content (magic bytes)
func (v *FileValidator) ValidateContent(reader io.Reader, expectedType string) error {
	// Read first 512 bytes for magic number detection
	buffer := make([]byte, 512)
	n, err := reader.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	// Check magic bytes
	contentType := v.detectContentType(buffer[:n])
	if contentType == "" {
		return fmt.Errorf("unable to detect file type from content")
	}

	// Validate against expected type
	if expectedType != "" && contentType != expectedType {
		return fmt.Errorf("file content type %s does not match expected type %s", contentType, expectedType)
	}

	return nil
}

// ValidateFile kiểm tra file hoàn chỉnh
func (v *FileValidator) ValidateFile(filename, contentType string, size int64, reader io.Reader, category string) error {
	// Validate file type
	allowedTypes, exists := v.allowedTypes[category]
	if !exists {
		allowedTypes = v.allowedTypes["default"]
	}

	if err := v.ValidateType(contentType, allowedTypes); err != nil {
		return err
	}

	// Validate file size
	maxSize, exists := v.maxSizes[category]
	if !exists {
		maxSize = v.maxSizes["default"]
	}

	if err := v.ValidateSize(size, maxSize); err != nil {
		return err
	}

	// Validate content if reader is provided
	if reader != nil {
		if err := v.ValidateContent(reader, contentType); err != nil {
			return err
		}
	}

	return nil
}

// SetAllowedTypes thiết lập allowed types cho category
func (v *FileValidator) SetAllowedTypes(category string, types []string) {
	v.allowedTypes[category] = types
}

// SetMaxSize thiết lập max size cho category
func (v *FileValidator) SetMaxSize(category string, size int64) {
	v.maxSizes[category] = size
}

// GetAllowedTypes lấy allowed types cho category
func (v *FileValidator) GetAllowedTypes(category string) []string {
	types, exists := v.allowedTypes[category]
	if !exists {
		return v.allowedTypes["default"]
	}
	return types
}

// GetMaxSize lấy max size cho category
func (v *FileValidator) GetMaxSize(category string) int64 {
	size, exists := v.maxSizes[category]
	if !exists {
		return v.maxSizes["default"]
	}
	return size
}

// detectContentType detect content type từ magic bytes
func (v *FileValidator) detectContentType(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Check common file signatures
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "image/jpeg"
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "image/png"
	case bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46, 0x38}):
		return "image/gif"
	case bytes.HasPrefix(data, []byte{0x42, 0x4D}):
		return "image/bmp"
	case bytes.HasPrefix(data, []byte{0x25, 0x50, 0x44, 0x46}):
		return "application/pdf"
	case bytes.HasPrefix(data, []byte{0x50, 0x4B, 0x03, 0x04}):
		return "application/zip"
	case bytes.HasPrefix(data, []byte{0x52, 0x61, 0x72, 0x21}):
		return "application/x-rar-compressed"
	case bytes.HasPrefix(data, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}):
		return "application/x-7z-compressed"
	case bytes.HasPrefix(data, []byte{0x1F, 0x8B}):
		return "application/gzip"
	case bytes.HasPrefix(data, []byte{0x49, 0x44, 0x33}):
		return "audio/mpeg"
	case bytes.HasPrefix(data, []byte{0xFF, 0xFB}):
		return "audio/mpeg"
	case bytes.HasPrefix(data, []byte{0x52, 0x49, 0x46, 0x46}):
		// Check for WAV or AVI
		if len(data) >= 12 && bytes.Equal(data[8:12], []byte{0x57, 0x41, 0x56, 0x45}) {
			return "audio/wav"
		}
		if len(data) >= 12 && bytes.Equal(data[8:12], []byte{0x41, 0x56, 0x49, 0x20}) {
			return "video/avi"
		}
		return "application/octet-stream"
	case bytes.HasPrefix(data, []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x32}):
		return "video/mp4"
	case bytes.HasPrefix(data, []byte{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x32}):
		return "video/mp4"
	}

	// Fallback to mime type detection
	contentType := mime.TypeByExtension(filepath.Ext(string(data)))
	if contentType != "" {
		return contentType
	}

	return "application/octet-stream"
}

// IsImage kiểm tra file có phải ảnh không
func (v *FileValidator) IsImage(contentType string) bool {
	allowedTypes := v.GetAllowedTypes("image")
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}

// IsDocument kiểm tra file có phải document không
func (v *FileValidator) IsDocument(contentType string) bool {
	allowedTypes := v.GetAllowedTypes("document")
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}

// IsVideo kiểm tra file có phải video không
func (v *FileValidator) IsVideo(contentType string) bool {
	allowedTypes := v.GetAllowedTypes("video")
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}

// IsAudio kiểm tra file có phải audio không
func (v *FileValidator) IsAudio(contentType string) bool {
	allowedTypes := v.GetAllowedTypes("audio")
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return true
		}
	}
	return false
}
