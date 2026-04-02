package entity

import (
	"time"

	"github.com/google/uuid"
)

type ReservationStatus string

const (
	ReservationPending   ReservationStatus = "pending"
	ReservationActive    ReservationStatus = "active"
	ReservationCompleted ReservationStatus = "completed"
	ReservationCancelled ReservationStatus = "cancelled"
)

type Reservation struct {
	ID            uuid.UUID         `db:"id"`
	UserID        uuid.UUID         `db:"user_id"`
	LibraryBookID uuid.UUID         `db:"library_book_id"`
	Status        ReservationStatus `db:"status"`
	ReservedAt    time.Time         `db:"reserved_at"`
	DueDate       time.Time         `db:"due_date"`
	ReturnedAt    *time.Time        `db:"returned_at"`
}

// ReservationView — read model for GET endpoints.
type ReservationView struct {
	ID            uuid.UUID         `db:"id"`
	Status        ReservationStatus `db:"status"`
	ReservedAt    time.Time         `db:"reserved_at"`
	DueDate       time.Time         `db:"due_date"`
	ReturnedAt    *time.Time        `db:"returned_at"`
	UserID        uuid.UUID         `db:"user_id"`
	UserName      string            `db:"user_name"`
	UserSurname   string            `db:"user_surname"`
	UserEmail     string            `db:"user_email"`
	LibraryBookID uuid.UUID         `db:"library_book_id"`
	BookID        uuid.UUID         `db:"book_id"`
	BookTitle     string            `db:"book_title"`
	BookCoverURL  string            `db:"book_cover_url"`
	LibraryID     uuid.UUID         `db:"library_id"`
	LibraryName   string            `db:"library_name"`
}
