package upload

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func uploadWorker(ctx context.Context, storage StorageClient, fileInput FileInput, bucket string, resultsCh chan<- UploadResult) {
	fileID := uuid.New().String()
	start := time.Now()

	objectKey, err := storage.Upload(ctx, fileInput, bucket)

	resultsCh <- UploadResult{
		FileID:      fileID,
		FileName:    fileInput.Filename,
		ObjectKey:   objectKey,
		ContentType: fileInput.ContentType,
		Size:        fileInput.Size,
		Error:       err,
		Duration:    time.Since(start),
	}
}
