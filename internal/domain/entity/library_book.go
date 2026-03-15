package entity

import "github.com/google/uuid"

type LibraryBook struct {
	ID              uuid.UUID `db:"id"`
	LibraryID       uuid.UUID `db:"library_id"`
	BookID          uuid.UUID `db:"book_id"`
	TotalCopies     int       `db:"total_copies"`
	AvailableCopies int       `db:"available_copies"`
}
