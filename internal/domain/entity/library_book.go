package entity

import "github.com/google/uuid"

type LibraryBook struct {
	ID              uuid.UUID `db:"id"`
	LibraryID       uuid.UUID `db:"library_id"`
	BookID          uuid.UUID `db:"book_id"`
	TotalCopies     int       `db:"total_copies"`
	AvailableCopies int       `db:"available_copies"`
}

// LibraryBookView — read model for GET endpoints.
// Includes library name and full book info (title, cover, author, genre).
type LibraryBookView struct {
	ID              uuid.UUID `db:"id"`
	LibraryID       uuid.UUID `db:"library_id"`
	LibraryName     string    `db:"library_name"`
	BookID          uuid.UUID `db:"book_id"`
	BookTitle       string    `db:"book_title"`
	BookCoverURL    string    `db:"book_cover_url"`
	BookYear        int       `db:"book_year"`
	AuthorID        uuid.UUID `db:"author_id"`
	AuthorName      string    `db:"author_name"`
	GenreID         uuid.UUID `db:"genre_id"`
	GenreName       string    `db:"genre_name"`
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
