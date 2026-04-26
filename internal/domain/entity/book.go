package entity

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID            uuid.UUID  `db:"id"`
	AuthorID      uuid.UUID  `db:"author_id"`
	GenreID       uuid.UUID  `db:"genre_id"`
	Title         string     `db:"title"`
	ISBN          string     `db:"isbn"`
	CoverURL      string     `db:"cover_url"`
	Description   string     `db:"description"`
	DescriptionKK string     `db:"description_kk"`
	Language      string     `db:"language"`
	Year          int        `db:"year"`
	AvgRating     float64    `db:"avg_rating"`
	TotalPages    *int       `db:"total_pages"`
	CreatedAt     time.Time  `db:"created_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

// BookView — read model for GET endpoints (Book + joined author/genre names).
// Populated via JOIN; write ops use Book directly.
type BookView struct {
	ID uuid.UUID `db:"id"`

	AuthorID        uuid.UUID `db:"author_id"`
	AuthorName      string    `db:"author_name"`
	AuthorBio       string    `db:"author_bio"`
	AuthorBioKK     string    `db:"author_bio_kk"`
	AuthorBirthDate *string   `db:"author_birth_date"` // "2006-01-02"
	AuthorDeathDate *string   `db:"author_death_date"` // "2006-01-02"
	AuthorPhotoURL  string    `db:"author_photo_url"`

	GenreID   uuid.UUID `db:"genre_id"`
	GenreName string    `db:"genre_name"`
	GenreSlug string    `db:"genre_slug"`

	BookFileID     *uuid.UUID      `db:"book_file_id"`
	BookFileBookID *uuid.UUID      `db:"book_file_book_id"`
	BookFileFormat *BookFileFormat `db:"book_file_format"`
	BookFileUrl    *string         `db:"book_file_url"`

	Title         string    `db:"title"`
	ISBN          string    `db:"isbn"`
	CoverURL      string    `db:"cover_url"`
	Description   string    `db:"description"`
	DescriptionKK string    `db:"description_kk"`
	Language      string    `db:"language"`
	Year          int       `db:"year"`
	AvgRating     float64   `db:"avg_rating"`
	TotalPages    *int      `db:"total_pages"`
	CreatedAt     time.Time `db:"created_at"`
}
