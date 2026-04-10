package entity

import (
	"time"

	"github.com/google/uuid"
)

type ReadingNote struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	BookID    uuid.UUID `db:"book_id"`
	Position  string    `db:"position"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// ReadingNoteView — read model for GET endpoints.
type ReadingNoteView struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	BookID    uuid.UUID `db:"book_id"`
	BookTitle string    `db:"book_title"`
	Position  string    `db:"position"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
