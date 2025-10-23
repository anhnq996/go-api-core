package aws

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"anhnq/api-core/pkg/storage/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implementation cho AWS S3
type S3Storage struct {
	client        *s3.Client
	bucket        string
	region        string
	baseURL       string
	presignClient *s3.PresignClient
}

// S3Config cấu hình cho S3 storage
type S3Config struct {
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	BaseURL         string `json:"base_url"` // Custom base URL (optional)
}

// NewS3Storage tạo instance mới của S3Storage
func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	// Tạo AWS config
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Tạo S3 client
	client := s3.NewFromConfig(awsConfig)
	presignClient := s3.NewPresignClient(client)

	// Generate base URL nếu không được cung cấp
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, cfg.Region)
	}

	return &S3Storage{
		client:        client,
		bucket:        cfg.Bucket,
		region:        cfg.Region,
		baseURL:       baseURL,
		presignClient: presignClient,
	}, nil
}

// Upload file từ io.Reader
func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader, options *interfaces.UploadOptions) (*interfaces.FileInfo, error) {
	// Prepare input
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	}

	// Set content type
	if options != nil && options.ContentType != "" {
		input.ContentType = aws.String(options.ContentType)
	}

	// Set ACL
	if options != nil && options.ACL != "" {
		input.ACL = types.ObjectCannedACL(options.ACL)
	} else {
		input.ACL = types.ObjectCannedACLBucketOwnerFullControl
	}

	// Set cache control
	if options != nil && options.CacheControl != "" {
		input.CacheControl = aws.String(options.CacheControl)
	}

	// Set metadata
	if options != nil && options.Metadata != nil {
		input.Metadata = options.Metadata
	}

	// Upload file
	result, err := s.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Get file info
	info, err := s.GetInfo(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Set ETag from result
	if result.ETag != nil {
		info.ETag = *result.ETag
	}

	return info, nil
}

// UploadBytes upload file từ bytes
func (s *S3Storage) UploadBytes(ctx context.Context, key string, data []byte, options *interfaces.UploadOptions) (*interfaces.FileInfo, error) {
	return s.Upload(ctx, key, strings.NewReader(string(data)), options)
}

// Download file về io.Reader
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return result.Body, nil
}

// DownloadBytes download file về bytes
func (s *S3Storage) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	reader, err := s.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// GetInfo lấy thông tin file
func (s *S3Storage) GetInfo(ctx context.Context, key string) (*interfaces.FileInfo, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Generate URL
	fileURL := s.generateURL(key)

	// Convert metadata
	metadata := make(map[string]string)
	if result.Metadata != nil {
		for k, v := range result.Metadata {
			metadata[k] = v
		}
	}

	return &interfaces.FileInfo{
		Name:         filepath.Base(key),
		Size:         *result.ContentLength,
		ContentType:  aws.ToString(result.ContentType),
		Path:         key,
		URL:          fileURL,
		ETag:         aws.ToString(result.ETag),
		LastModified: result.LastModified.Unix(),
		Metadata:     metadata,
	}, nil
}

// Exists kiểm tra file có tồn tại không
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.GetInfo(ctx, key)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "NoSuchKey") || strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Delete xóa file
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteMultiple xóa nhiều files
func (s *S3Storage) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Convert keys to object identifiers
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	}

	_, err := s.client.DeleteObjects(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return nil
}

// List list files
func (s *S3Storage) List(ctx context.Context, options *interfaces.ListOptions) (*interfaces.ListResult, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	}

	// Set prefix
	if options != nil && options.Prefix != "" {
		input.Prefix = aws.String(options.Prefix)
	}

	// Set delimiter
	if options != nil && options.Delimiter != "" {
		input.Delimiter = aws.String(options.Delimiter)
	}

	// Set max keys
	if options != nil && options.MaxKeys > 0 {
		input.MaxKeys = aws.Int32(int32(options.MaxKeys))
	}

	// Set marker
	if options != nil && options.Marker != "" {
		input.ContinuationToken = aws.String(options.Marker)
	}

	result, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Convert files
	var files []interfaces.FileInfo
	for _, obj := range result.Contents {
		contentType := "application/octet-stream"
		// Note: S3 ListObjectsV2 doesn't return ContentType, need to use HeadObject for that

		files = append(files, interfaces.FileInfo{
			Name:         filepath.Base(*obj.Key),
			Size:         *obj.Size,
			ContentType:  contentType,
			Path:         *obj.Key,
			URL:          s.generateURL(*obj.Key),
			ETag:         *obj.ETag,
			LastModified: obj.LastModified.Unix(),
			Metadata:     make(map[string]string),
		})
	}

	// Convert folders (common prefixes)
	var folders []string
	for _, prefix := range result.CommonPrefixes {
		folders = append(folders, *prefix.Prefix)
	}

	return &interfaces.ListResult{
		Files:       files,
		Folders:     folders,
		NextMarker:  aws.ToString(result.NextContinuationToken),
		IsTruncated: aws.ToBool(result.IsTruncated),
	}, nil
}

// GetURL lấy public URL
func (s *S3Storage) GetURL(ctx context.Context, key string) (string, error) {
	return s.generateURL(key), nil
}

// GetSignedURL lấy signed URL
func (s *S3Storage) GetSignedURL(ctx context.Context, key string, expiresIn int64) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	presignResult, err := s.presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expiresIn) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return presignResult.URL, nil
}

// Copy copy file
func (s *S3Storage) Copy(ctx context.Context, srcKey, dstKey string) error {
	source := fmt.Sprintf("%s/%s", s.bucket, srcKey)

	input := &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String(url.QueryEscape(source)),
	}

	_, err := s.client.CopyObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// Move move file (copy + delete)
func (s *S3Storage) Move(ctx context.Context, srcKey, dstKey string) error {
	if err := s.Copy(ctx, srcKey, dstKey); err != nil {
		return err
	}

	return s.Delete(ctx, srcKey)
}

// generateURL tạo URL cho file
func (s *S3Storage) generateURL(key string) string {
	if !strings.HasSuffix(s.baseURL, "/") {
		return s.baseURL + "/" + key
	}
	return s.baseURL + key
}
