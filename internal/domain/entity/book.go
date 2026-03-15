package entity

import "github.com/google/uuid"

type Book struct {
	ID          uuid.UUID `db:"id"`
	AuthorID    uuid.UUID `db:"author_id"`
	GenreID     uuid.UUID `db:"genre_id"`
	Title       string    `db:"title"`
	ISBN        string    `db:"isbn"`
	CoverURL    string    `db:"cover_url"`
	Description string    `db:"description"`
	Language    string    `db:"language"`
	Year        int       `db:"year"`
	AvgRating   float64   `db:"avg_rating"`
}
