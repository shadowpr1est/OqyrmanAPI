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
	ID             uuid.UUID `db:"id"`
	LibraryID      uuid.UUID `db:"library_id"`
	LibraryName    string    `db:"library_name"`
	LibraryAddress string    `db:"library_address"`
	LibraryLat     float64   `db:"library_lat"`
	LibraryLng     float64   `db:"library_lng"`
	LibraryPhone   string    `db:"library_phone"`

	BookID          uuid.UUID `db:"book_id"`
	BookTitle       string    `db:"book_title"`
	BookISBN        string    `db:"book_isbn"`
	BookCoverURL    string    `db:"book_cover_url"`
	BookYear        int       `db:"book_year"`
	BookDescription string    `db:"book_description"`
	BookLanguage    string    `db:"book_language"`
	BookTotalPages  *int      `db:"book_total_pages"`
	BookAvgRating   float64   `db:"book_avg_rating"`

	AuthorID        uuid.UUID `db:"author_id"`
	AuthorName      string    `db:"author_name"`
	AuthorBio       string    `db:"author_bio"`
	AuthorBirthDate *string   `db:"author_birth_date"` // "2006-01-02"
	AuthorDeathDate *string   `db:"author_death_date"` // "2006-01-02"
	AuthorPhotoURL  string    `db:"author_photo_url"`

	GenreID   uuid.UUID `db:"genre_id"`
	GenreName string    `db:"genre_name"`
	GenreSlug string    `db:"genre_slug"`

	TotalCopies     int `db:"total_copies"`
	AvailableCopies int `db:"available_copies"`
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
