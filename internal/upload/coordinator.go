package upload

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

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

func (uc *UploadCoordinator) ProcessDeleteExpiredFiles() {
	files, err := uc.Repository.GetExpiredFiles()
	if err != nil {
		uc.Logger.Error("failed to get expired files",
			slog.String("error", err.Error()),
		)
		return
	}

	if len(files) == 0 {
		return
	}

	ctx := context.Background()

	uc.Logger.Info("starting expired file cleanup", slog.Int("count", len(files)))

	for _, f := range files {
		err := uc.Storage.Delete(ctx, uc.BucketName, f.ObjectKey)
		if err != nil {
			uc.Logger.Error("failed to delete file from storage",
				slog.String("file_id", f.FileID.String()),
				slog.String("filename", f.OriginalFilename),
				slog.String("object_key", f.ObjectKey),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := uc.Repository.DeleteFileFromDB(ctx, f.FileID); err != nil {
			uc.Logger.Error("failed to delete file from DB",
				slog.String("file_id", f.FileID.String()),
				slog.String("object_key", f.ObjectKey),
				slog.String("error", err.Error()),
			)
		}
	}

	uc.Logger.Info("expired file cleanup completed", slog.Int("count", len(files)))
}

func (uc *UploadCoordinator) StartCleanupScheduler(ctx context.Context, intervalMinutes int) {
	interval := time.Duration(intervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	uc.Logger.Info("cleanup scheduler started", slog.Int("interval_minutes", intervalMinutes))

	for {
		select {
		case <-ticker.C:
			uc.Logger.Info("running scheduled cleanup")
			uc.ProcessDeleteExpiredFiles()
		case <-ctx.Done():
			uc.Logger.Info("cleanup scheduler stopped")
			return
		}
	}
}
