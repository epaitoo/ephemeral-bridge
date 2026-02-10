package data

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type File struct {
	ID               uuid.UUID
	Bucket           string
	ObjectKey        string
	OriginalFilename *string
	ContentType      *string
	SizeBytes        int64
	Checksum         *string
	CreatedAt        time.Time
	ExpiresAt        *time.Time
	DownloadedAt     *time.Time
}

type FileResponse struct {
	ID        uuid.UUID `json:"id"`
	Filename  *string   `json:"filename"`
	SizeBytes int64     `json:"size_bytes"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (f *File) ToResponse() FileResponse {
	return FileResponse{
		ID:        f.ID,
		Filename:  f.OriginalFilename,
		SizeBytes: f.SizeBytes,
		ExpiresAt: f.ExpiresAt,
	}
}

type FileModel struct {
	DB *pgxpool.Pool
}

func (f FileModel) Insert(file *File) error {
	query := `INSERT INTO files (bucket, object_key, original_filename, content_type, size_bytes, checksum)
					VALUES ($1, $2, $3, $4, $5, $6)
					RETURNING id, original_filename, size_bytes, expires_at
	`

	args := []any{file.Bucket, file.ObjectKey, file.OriginalFilename, file.ContentType, file.SizeBytes, file.Checksum}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return f.DB.QueryRow(ctx, query, args...).Scan(&file.ID, &file.OriginalFilename, &file.SizeBytes,
		&file.ExpiresAt)
}

func (f FileModel) GetAllFiles() (*[]File, error) {
	query := `
		SELECT id, original_filename, size_bytes, expires_at 
		FROM files`

	var files []File

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := f.DB.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var file File

		err := rows.Scan(
			&file.ID,
			&file.OriginalFilename,
			&file.SizeBytes,
			&file.ExpiresAt,
		)

		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	// Check for errors encountered during iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &files, nil

}

func (f FileModel) Get(id uuid.UUID) (*File, error) {
	query := `
		SELECT id, original_filename, size_bytes, expires_at 
		FROM files
		WHERE id = $1`

	var file File

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := f.DB.QueryRow(ctx, query, id).Scan(
		&file.ID,
		&file.OriginalFilename,
		&file.SizeBytes,
		&file.ExpiresAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &file, nil
}

func (f FileModel) Delete(id uuid.UUID) error {
	query := `
		DELETE FROM files
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := f.DB.Exec(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
