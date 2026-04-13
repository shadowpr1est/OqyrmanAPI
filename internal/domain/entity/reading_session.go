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
	Progress    int           `db:"progress"`
	CfiPosition *string       `db:"cfi_position"`
	Status      ReadingStatus `db:"status"`
	UpdatedAt   time.Time     `db:"updated_at"`
	FinishedAt  *time.Time    `db:"finished_at"`
}

// ReadingSessionView — read model for GET endpoints.
type ReadingSessionView struct {
	ID           uuid.UUID     `db:"id"`
	UserID       uuid.UUID     `db:"user_id"`
	BookID       uuid.UUID     `db:"book_id"`
	BookTitle    string        `db:"book_title"`
	BookCoverURL string        `db:"book_cover_url"`
	AuthorID     uuid.UUID     `db:"author_id"`
	AuthorName   string        `db:"author_name"`
	Progress     int           `db:"progress"`
	CfiPosition  *string       `db:"cfi_position"`
	Status       ReadingStatus `db:"status"`
	UpdatedAt    time.Time     `db:"updated_at"`
	FinishedAt   *time.Time    `db:"finished_at"`
}