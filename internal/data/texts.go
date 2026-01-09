package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Texts struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	TextDetails string    `json:"text_details"`
	CreatedAt   time.Time `json:"created_At"`
	UpdatedAt   time.Time `json:"updated_At"`
}

type TextModel struct {
	DB *pgxpool.Pool
}

func (t TextModel) GetAllTexts() (*[]Texts, error) {
	query := `
		SELECT id, name, text_details, created_at, updated_at 
		FROM texts`

	var texts []Texts

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := t.DB.Query(ctx, query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var text Texts

		err := rows.Scan(
			&text.ID,
			&text.Name,
			&text.TextDetails,
			&text.CreatedAt,
			&text.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		texts = append(texts, text)
	}

	// Check for errors encountered during iteration.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &texts, nil
}

func (t TextModel) Insert(text *Texts) error {
	query := `INSERT INTO texts (name, text_details)
					VALUES ($1, $2)
					RETURNING id, name, text_details, created_at, updated_at
	`
	args := []any{text.Name, text.TextDetails}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return t.DB.QueryRow(ctx, query, args...).Scan(&text.ID, &text.Name, &text.TextDetails,
		&text.CreatedAt, &text.UpdatedAt)
}

func (t *TextModel) Get(id int64) (*Texts, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, text_details, created_at, updated_at 
		FROM texts
		WHERE id = $1`

	var text Texts

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := t.DB.QueryRow(ctx, query, id).Scan(
		&text.ID,
		&text.Name,
		&text.TextDetails,
		&text.CreatedAt,
		&text.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &text, nil
}

func (t *TextModel) Update(text *Texts) error {
	query := `
		UPDATE texts
		SET name = $1, text_details = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, text_details, created_at, updated_at
	`
	args := []any{
		text.Name,
		text.TextDetails,
		text.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return t.DB.QueryRow(ctx, query, args...).Scan(
		&text.ID,
		&text.Name,
		&text.TextDetails,
		&text.CreatedAt,
		&text.UpdatedAt,
	)
}

func (t *TextModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM texts
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := t.DB.Exec(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}
