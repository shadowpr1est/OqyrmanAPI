package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/shadowpr1est/OqyrmanAPI/config"
)

type MinioStorage struct {
	client   *minio.Client
	bucket   string
	endpoint string
	useSSL   bool
}

func NewMinioStorage(cfg *config.Config) (*MinioStorage, error) {
	client, err := minio.New(cfg.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Minio.AccessKey, cfg.Minio.SecretKey, ""),
		Secure: cfg.Minio.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio init: %w", err)
	}

	// создать бакет если нет
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Minio.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Minio.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio make bucket: %w", err)
		}
	}

	return &MinioStorage{
		client:   client,
		bucket:   cfg.Minio.Bucket,
		endpoint: cfg.Minio.Endpoint,
		useSSL:   cfg.Minio.UseSSL,
	}, nil
}

func (s *MinioStorage) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio upload: %w", err)
	}

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, objectKey)
	return url, nil
}

func (s *MinioStorage) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
}

func (s *MinioStorage) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	return err
}
