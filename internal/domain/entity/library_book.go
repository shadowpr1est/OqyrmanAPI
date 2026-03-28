package entity

import "github.com/google/uuid"

type LibraryBook struct {
	ID              uuid.UUID `db:"id"`
	LibraryID       uuid.UUID `db:"library_id"`
	BookID          uuid.UUID `db:"book_id"`
	TotalCopies     int       `db:"total_copies"`
	AvailableCopies int       `db:"available_copies"`
}

type LibraryBookSearchResult struct {
	LibraryBookID   uuid.UUID `db:"library_book_id"`
	BookID          uuid.UUID `db:"book_id"`
	Title           string    `db:"title"`
	Author          string    `db:"author"`
	Genre           string    `db:"genre"`
	CoverURL        *string   `db:"cover_url"`
	Year            *int      `db:"year"`
	TotalCopies     int       `db:"total_copies"`
	AvailableCopies int       `db:"available_copies"`
	IsAvailable     bool      `db:"is_available"`
}
