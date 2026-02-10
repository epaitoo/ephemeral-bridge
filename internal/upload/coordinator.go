package upload

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type UploadCoordinator struct {
	Storage    StorageClient
	Repository FileRepository
	Logger     *slog.Logger
	BucketName string
}

func (uc *UploadCoordinator) ProcessUploads(files []FileInput) {
	resultsCh := make(chan UploadResult, len(files))
	var wg sync.WaitGroup

	ctx := context.Background()

	uc.Logger.Info("starting upload processing", slog.Int("count", len(files)))

	for _, f := range files {
		wg.Add(1)
		go func(fileInput FileInput) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					resultsCh <- UploadResult{
						FileID:   uuid.New().String(),
						FileName: fileInput.Filename,
						Error:    fmt.Errorf("worker panicked: %v", r),
					}
				}
			}()
			uploadWorker(ctx, uc.Storage, fileInput, uc.BucketName, resultsCh)
		}(f)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	for result := range resultsCh {
		if result.Error != nil {
			uc.Logger.Error("upload failed",
				slog.String("file_id", result.FileID),
				slog.String("filename", result.FileName),
				slog.Duration("duration", result.Duration),
				slog.String("error", result.Error.Error()),
			)
			continue
		}

		uc.Logger.Info("upload succeeded",
			slog.String("file_id", result.FileID),
			slog.String("filename", result.FileName),
			slog.String("object_key", result.ObjectKey),
			slog.Int64("size_bytes", result.Size),
			slog.Duration("duration", result.Duration),
		)

		metadata := FileMetadata{
			Bucket:           uc.BucketName,
			ObjectKey:        result.ObjectKey,
			OriginalFilename: result.FileName,
			ContentType:      result.ContentType,
			SizeBytes:        result.Size,
		}

		if err := uc.Repository.SaveFileMetadata(ctx, metadata); err != nil {
			uc.Logger.Error("DB insert failed, cleaning up R2 object",
				slog.String("file_id", result.FileID),
				slog.String("object_key", result.ObjectKey),
				slog.String("error", err.Error()),
			)
			uc.Storage.Delete(ctx, uc.BucketName, result.ObjectKey)
		}
	}

	uc.Logger.Info("all file uploads processed", slog.Int("count", len(files)))
}
