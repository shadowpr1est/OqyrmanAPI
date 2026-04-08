package entity

import (
	"time"

	"github.com/google/uuid"
)

type ShelfStatus string

const (
	ShelfWantToRead ShelfStatus = "want_to_read"
	ShelfReading    ShelfStatus = "reading"
	ShelfFinished   ShelfStatus = "finished"
)

type Wishlist struct {
	ID      uuid.UUID   `db:"id"`
	UserID  uuid.UUID   `db:"user_id"`
	BookID  uuid.UUID   `db:"book_id"`
	Status  ShelfStatus `db:"status"`
	AddedAt time.Time   `db:"added_at"`
}

// WishlistView — read model for GET /wishlist.
type WishlistView struct {
	ID            uuid.UUID   `db:"id"`
	UserID        uuid.UUID   `db:"user_id"`
	BookID        uuid.UUID   `db:"book_id"`
	BookTitle     string      `db:"book_title"`
	BookCoverURL  string      `db:"book_cover_url"`
	BookAvgRating float64     `db:"book_avg_rating"`
	AuthorID      uuid.UUID   `db:"author_id"`
	AuthorName    string      `db:"author_name"`
	GenreID       uuid.UUID   `db:"genre_id"`
	GenreName     string      `db:"genre_name"`
	Status        ShelfStatus `db:"status"`
	AddedAt       time.Time   `db:"added_at"`
}
