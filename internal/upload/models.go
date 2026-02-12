package upload

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type UploadResult struct {
	FileID      string
	FileName    string
	ObjectKey   string
	ContentType string
	Size        int64
	Error       error
	Duration    time.Duration
}

type FileInput struct {
	Reader      io.ReadSeeker
	Filename    string
	ContentType string
	Size        int64
}

type FileMetadata struct {
	FileID           uuid.UUID
	Bucket           string
	ObjectKey        string
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
}
