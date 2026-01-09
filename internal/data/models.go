package data

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is returned when a there is a data race, and we have an edit conflict.
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	Texts TextModel
}

func NewModels(db *pgxpool.Pool) Models {
	return Models{
		Texts: TextModel{DB: db},
	}
}
