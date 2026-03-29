package entity

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID          uuid.UUID  `db:"id"`
	AuthorID    uuid.UUID  `db:"author_id"`
	GenreID     uuid.UUID  `db:"genre_id"`
	Title       string     `db:"title"`
	ISBN        string     `db:"isbn"`
	CoverURL    string     `db:"cover_url"`
	Description string     `db:"description"`
	Language    string     `db:"language"`
	Year        int        `db:"year"`
	AvgRating   float64    `db:"avg_rating"`
	TotalPages  *int       `db:"total_pages"`
	CreatedAt   time.Time  `db:"created_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

// BookView — read model for GET endpoints (Book + joined author/genre names).
// Populated via JOIN; write ops use Book directly.
type BookView struct {
	ID          uuid.UUID `db:"id"`
	AuthorID    uuid.UUID `db:"author_id"`
	AuthorName  string    `db:"author_name"`
	GenreID     uuid.UUID `db:"genre_id"`
	GenreName   string    `db:"genre_name"`
	Title       string    `db:"title"`
	ISBN        string    `db:"isbn"`
	CoverURL    string    `db:"cover_url"`
	Description string    `db:"description"`
	Language    string    `db:"language"`
	Year        int       `db:"year"`
	AvgRating   float64   `db:"avg_rating"`
	TotalPages  *int      `db:"total_pages"`
	CreatedAt   time.Time `db:"created_at"`
}
