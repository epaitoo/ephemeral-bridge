package upload

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type StorageClient interface {
	Upload(ctx context.Context, file FileInput, bucket string) (objectKey string, err error)
	Delete(ctx context.Context, bucket, objectKey string) error
	GenerateDownloadURL(ctx context.Context, bucket, objectKey string, expiry time.Duration) (string, error)
}

type FileRepository interface {
	SaveFileMetadata(ctx context.Context, metadata FileMetadata) error
	GetExpiredFiles() ([]FileMetadata, error)
	DeleteFileFromDB(ctx context.Context, id uuid.UUID) error
}
