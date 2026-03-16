package storage

import (
	"context"
	"io"
)

type FileStorage interface {
	Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, objectKey string) error
}
