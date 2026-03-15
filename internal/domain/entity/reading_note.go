package entity

import (
	"time"

	"github.com/google/uuid"
)

type ReadingNote struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	BookID    uuid.UUID `db:"book_id"`
	Page      int       `db:"page"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}
