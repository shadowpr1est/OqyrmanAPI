package entity

import (
	"time"

	"github.com/google/uuid"
)

type ReadingStatus string

const (
	StatusReading  ReadingStatus = "reading"
	StatusFinished ReadingStatus = "finished"
	StatusDropped  ReadingStatus = "dropped"
)

type ReadingSession struct {
	ID          uuid.UUID     `db:"id"`
	UserID      uuid.UUID     `db:"user_id"`
	BookID      uuid.UUID     `db:"book_id"`
	CurrentPage int           `db:"current_page"`
	Status      ReadingStatus `db:"status"`
	UpdatedAt   time.Time     `db:"updated_at"`
}
