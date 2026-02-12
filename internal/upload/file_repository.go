package upload

import (
	"context"

	"github.com/epaitoo/ephermalbridge/internal/data"
	"github.com/google/uuid"
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

func (r *PGFileRepository) GetExpiredFiles() ([]FileMetadata, error) {
	files, err := r.Model.GetExpiredFiles()
	if err != nil {
		return nil, err
	}

	var filesMetaData []FileMetadata

	for _, file := range files {
		var fileMetaData FileMetadata

		fileMetaData.Bucket = file.Bucket
		fileMetaData.ObjectKey = file.ObjectKey
		if file.OriginalFilename != nil {
			fileMetaData.OriginalFilename = *file.OriginalFilename
		}
		fileMetaData.FileID = file.ID

		filesMetaData = append(filesMetaData, fileMetaData)
	}

	return filesMetaData, nil
}

// Delete file from DB
func (r *PGFileRepository) DeleteFileFromDB(_ context.Context, id uuid.UUID) error {
	return r.Model.Delete(id)
}
