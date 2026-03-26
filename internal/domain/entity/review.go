package entity

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	BookID    uuid.UUID  `db:"book_id"`
	Rating    int        `db:"rating"` // 1-5
	Body      string     `db:"body"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}
