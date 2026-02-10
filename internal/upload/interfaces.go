package upload

import "context"

type StorageClient interface {
	Upload(ctx context.Context, file FileInput, bucket string) (objectKey string, err error)
	Delete(ctx context.Context, bucket, objectKey string) error
}

type FileRepository interface {
	SaveFileMetadata(ctx context.Context, metadata FileMetadata) error
}
