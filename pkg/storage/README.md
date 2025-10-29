# Storage Module

Storage module cung cấp khả năng upload, quản lý và xử lý files với hỗ trợ nhiều storage backend.

## Tính năng

### 🚀 Core Features

- **Multi-backend Support**: Local storage và AWS S3
- **File Upload**: Upload files với validation
- **Image Processing**: Resize, crop, watermark, format conversion
- **File Management**: List, delete, copy, move files
- **Security**: File validation, size limits, type checking
- **Metadata**: Custom metadata support
- **Public/Private**: Control file access

### 📁 Supported File Types

- **Images**: JPEG, PNG, GIF, WebP, BMP, TIFF
- **Documents**: PDF, DOC, DOCX, XLS, XLSX, PPT, PPTX, TXT, CSV
- **Videos**: MP4, AVI, MOV, WMV, FLV, WebM, MKV
- **Audio**: MP3, WAV, OGG, AAC, FLAC
- **Archives**: ZIP, RAR, 7Z, GZIP, TAR

## Cấu hình

### Environment Variables

```bash
# Storage Driver (local, s3)
STORAGE_DRIVER=local

# Local Storage
STORAGE_LOCAL_PATH=storage/app
STORAGE_LOCAL_URL=/storage

# AWS S3
STORAGE_S3_BUCKET=your-bucket-name
STORAGE_S3_REGION=us-east-1
STORAGE_S3_ACCESS_KEY_ID=your-access-key
STORAGE_S3_SECRET_ACCESS_KEY=your-secret-key
STORAGE_S3_BASE_URL=https://your-bucket.s3.region.amazonaws.com

# Image Processing
STORAGE_IMAGE_QUALITY=90

# File Validation
STORAGE_MAX_FILE_SIZE=10485760  # 10MB
```

## Sử dụng

### 1. Khởi tạo Storage

```go
import (
    "api-core/config"
    "api-core/pkg/storage"
)

// Lấy config
cfg := config.GetDefaultStorageConfig()

// Tạo storage factory
factory := storage.NewStorageFactory()

// Tạo storage components
storage, imageProcessor, validator, err := factory.CreateStorageComponents(cfg)
if err != nil {
    log.Fatal(err)
}
```

### 2. Upload File

```go
// Upload file từ multipart form
file, err := storageService.UploadFile(ctx, fileHeader, userID, &storage.UploadOptions{
    Category:     "image",
    Path:         "uploads/images",
    Public:       true,
    ProcessImage: true,
    ImageOptions: &storage.ImageOptions{
        Resize: &storage.ResizeOptions{
            Width:  800,
            Height: 600,
        },
        Quality: 90,
    },
})
```

### 3. Image Processing

```go
// Resize ảnh
resizedReader, err := imageProcessor.Resize(ctx, originalReader, 800, 600)

// Crop ảnh
croppedReader, err := imageProcessor.Crop(ctx, originalReader, 100, 100, 400, 300)

// Thêm watermark
watermarkedReader, err := imageProcessor.AddWatermark(ctx, originalReader, "watermark.png", "bottom-right")

// Convert format
convertedReader, err := imageProcessor.Convert(ctx, originalReader, "jpeg")
```

### 4. File Management

```go
// List files
files, pagination, err := storageService.ListFiles(ctx, &storage.ListOptions{
    Page:     1,
    PerPage:  20,
    Category: "image",
    Search:   "profile",
})

// Get file info
file, err := storageService.GetFile(ctx, fileID)

// Download file
reader, file, err := storageService.DownloadFile(ctx, fileID)

// Delete file
err := storageService.DeleteFile(ctx, fileID)

// Copy file
copiedFile, err := storageService.CopyFile(ctx, fileID, "new/path/file.jpg")
```

## API Endpoints

### Upload File

```
POST /api/v1/storage/upload
Content-Type: multipart/form-data

Form fields:
- file: File to upload (required)
- category: File category (required)
- path: Custom path (optional)
- public: Public access (optional)
- process_image: Process image (optional)
```

### List Files

```
GET /api/v1/storage/files?page=1&per_page=20&category=image&search=profile
```

### Get File Info

```
GET /api/v1/storage/files/{id}
```

### Download File

```
GET /api/v1/storage/files/{id}/download
```

### Get File URL

```
GET /api/v1/storage/files/{id}/url?signed=true&expires_in=3600
```

### Copy File

```
POST /api/v1/storage/files/{id}/copy
{
    "new_path": "new/path/file.jpg"
}
```

### Delete File

```
DELETE /api/v1/storage/files/{id}
```

### Get Public Files

```
GET /api/v1/storage/public?page=1&per_page=20&category=image
```

## File Validation

### Automatic Validation

- **File Type**: Kiểm tra MIME type và magic bytes
- **File Size**: Kiểm tra kích thước file
- **File Content**: Validate nội dung file

### Custom Validation

```go
// Set custom allowed types
validator.SetAllowedTypes("image", []string{
    "image/jpeg",
    "image/png",
    "image/gif",
})

// Set custom max size
validator.SetMaxSize("image", 5*1024*1024) // 5MB
```

## Image Processing Options

### Resize Options

```go
resizeOptions := &storage.ResizeOptions{
    Width:  800,
    Height: 600,
}
```

### Crop Options

```go
cropOptions := &storage.CropOptions{
    X:      100,
    Y:      100,
    Width:  400,
    Height: 300,
}
```

### Watermark Options

```go
watermarkOptions := &storage.WatermarkOptions{
    Path:     "watermarks/logo.png",
    Position: "bottom-right", // top-left, top-right, bottom-left, bottom-right, center
}
```

## Storage Backends

### Local Storage

- Lưu trữ files trong local file system
- Phù hợp cho development và small applications
- Cấu hình đơn giản

### AWS S3

- Lưu trữ files trên AWS S3
- Phù hợp cho production và large applications
- Hỗ trợ CDN và global distribution
- Signed URLs cho private files

## Security

### File Validation

- Kiểm tra file type bằng MIME type và magic bytes
- Giới hạn kích thước file
- Validate file content

### Access Control

- Public/Private files
- User-based access control
- Signed URLs cho private files

### Path Security

- Prevent directory traversal attacks
- Validate file paths
- Sanitize filenames

## Performance

### Optimization

- Lazy loading
- Efficient file streaming
- Image processing optimization
- Caching support

### Scalability

- Multiple storage backends
- CDN integration
- Load balancing support

## Error Handling

### Common Errors

- `FileNotFound`: File không tồn tại
- `InvalidFileType`: File type không được hỗ trợ
- `FileTooLarge`: File quá lớn
- `UploadFailed`: Upload thất bại
- `ProcessingFailed`: Xử lý file thất bại

### Error Response Format

```json
{
  "success": false,
  "code": "INVALID_FILE_TYPE",
  "message": "File type not supported",
  "errors": {
    "file": ["File type image/svg+xml is not allowed"]
  }
}
```

## Testing

### Unit Tests

```bash
go test ./pkg/storage/...
```

### Integration Tests

```bash
go test ./internal/app/storage/...
```

## Examples

### Complete Upload Example

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // Parse multipart form
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    // Get file
    file, fileHeader, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "No file provided", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Get user ID from context
    userID := r.Context().Value("user_id").(string)

    // Upload options
    options := &storage.UploadOptions{
        Category:     "image",
        Public:       true,
        ProcessImage: true,
        ImageOptions: &storage.ImageOptions{
            Resize: &storage.ResizeOptions{
                Width:  800,
                Height: 600,
            },
            Quality: 90,
        },
    }

    // Upload file
    uploadedFile, err := storageService.UploadFile(r.Context(), fileHeader, userID, options)
    if err != nil {
        http.Error(w, "Upload failed", http.StatusInternalServerError)
        return
    }

    // Return response
    json.NewEncoder(w).Encode(uploadedFile)
}
```

## Migration

### Database Migration

```sql
-- Create files table
CREATE TABLE files (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL UNIQUE,
    size BIGINT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    public BOOLEAN DEFAULT FALSE,
    url VARCHAR(500),
    etag VARCHAR(100),
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    INDEX idx_files_user_id (user_id),
    INDEX idx_files_category (category),
    INDEX idx_files_public (public),
    INDEX idx_files_created_at (created_at),
    INDEX idx_files_deleted_at (deleted_at),

    CONSTRAINT fk_files_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## Best Practices

### File Organization

- Sử dụng folder structure rõ ràng
- Đặt tên file có ý nghĩa
- Sử dụng timestamps cho unique names

### Security

- Validate tất cả file uploads
- Giới hạn file size và types
- Sử dụng signed URLs cho private files
- Sanitize file paths

### Performance

- Sử dụng CDN cho public files
- Optimize images trước khi upload
- Implement caching
- Sử dụng async processing cho large files

### Monitoring

- Log file operations
- Monitor storage usage
- Track upload/download metrics
- Alert on storage errors
