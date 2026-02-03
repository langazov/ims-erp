package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorageService implements domain.StorageService using MinIO
type MinIOStorageService struct {
	client  *minio.Client
	enpoint string
	useSSL  bool
}

// MinIOConfig holds configuration for MinIO connection
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
}

// NewMinIOStorageService creates a new MinIO storage service
func NewMinIOStorageService(config MinIOConfig) (*MinIOStorageService, error) {
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &MinIOStorageService{
		client:  minioClient,
		enpoint: config.Endpoint,
		useSSL:  config.UseSSL,
	}, nil
}

// Upload uploads data to MinIO storage
func (s *MinIOStorageService) Upload(ctx context.Context, bucket, objectKey string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)

	_, err := s.client.PutObject(ctx, bucket, objectKey, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

// Download retrieves data from MinIO storage
func (s *MinIOStorageService) Download(ctx context.Context, bucket, objectKey string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

// Delete removes an object from MinIO storage
func (s *MinIOStorageService) Delete(ctx context.Context, bucket, objectKey string) error {
	err := s.client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// GetPresignedUploadURL generates a presigned URL for uploading
func (s *MinIOStorageService) GetPresignedUploadURL(ctx context.Context, bucket, objectKey string, contentType string, expiry time.Duration) (string, error) {
	// Set default expiry
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	presignedURL, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return presignedURL.String(), nil
}

// GetPresignedDownloadURL generates a presigned URL for downloading
func (s *MinIOStorageService) GetPresignedDownloadURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error) {
	// Set default expiry
	if expiry <= 0 {
		expiry = 15 * time.Minute
	}

	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, objectKey, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	return presignedURL.String(), nil
}

// BucketExists checks if a bucket exists
func (s *MinIOStorageService) BucketExists(ctx context.Context, bucket string) (bool, error) {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	return exists, nil
}

// CreateBucket creates a new bucket
func (s *MinIOStorageService) CreateBucket(ctx context.Context, bucket string) error {
	err := s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}
