package upload

import (
	"context"

	"github.com/epaitoo/ephermalbridge/internal/data"
)

type PGFileRepository struct {
	Model data.FileModel
}

func (r *PGFileRepository) SaveFileMetadata(_ context.Context, metadata FileMetadata) error {
	file := &data.File{
		Bucket:           metadata.Bucket,
		ObjectKey:        metadata.ObjectKey,
		OriginalFilename: &metadata.OriginalFilename,
		ContentType:      &metadata.ContentType,
		SizeBytes:        metadata.SizeBytes,
	}

	return r.Model.Insert(file)
}
