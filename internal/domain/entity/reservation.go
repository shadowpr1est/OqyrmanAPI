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
	LibraryBookID uuid.UUID         `db:"library_book_id"` // NOT NULL — machine убран
	Status        ReservationStatus `db:"status"`
	ReservedAt    time.Time         `db:"reserved_at"`
	DueDate       time.Time         `db:"due_date"`
	ReturnedAt    *time.Time        `db:"returned_at"`
}
