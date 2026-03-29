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

// WishlistView — read model for GET /wishlist.
type WishlistView struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	BookID       uuid.UUID `db:"book_id"`
	BookTitle    string    `db:"book_title"`
	BookCoverURL string    `db:"book_cover_url"`
	BookAvgRating float64  `db:"book_avg_rating"`
	AuthorID     uuid.UUID `db:"author_id"`
	AuthorName   string    `db:"author_name"`
	GenreID      uuid.UUID `db:"genre_id"`
	GenreName    string    `db:"genre_name"`
	AddedAt      time.Time `db:"added_at"`
}
