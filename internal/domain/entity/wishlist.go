package entity

import (
	"time"

	"github.com/google/uuid"
)

type Wishlist struct {
	ID      uuid.UUID `db:"id"`
	UserID  uuid.UUID `db:"user_id"`
	BookID  uuid.UUID `db:"book_id"`
	AddedAt time.Time `db:"added_at"`
}
