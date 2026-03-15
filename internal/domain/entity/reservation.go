package entity

import (
	"time"

	"github.com/google/uuid"
)

type SourceType string
type ReservationStatus string

const (
	SourceLibrary SourceType = "library"
	SourceMachine SourceType = "machine"

	ReservationPending   ReservationStatus = "pending"
	ReservationActive    ReservationStatus = "active"
	ReservationCompleted ReservationStatus = "completed"
	ReservationCancelled ReservationStatus = "cancelled"
)

type Reservation struct {
	ID            uuid.UUID         `db:"id"`
	UserID        uuid.UUID         `db:"user_id"`
	LibraryBookID *uuid.UUID        `db:"library_book_id"` // nullable
	MachineBookID *uuid.UUID        `db:"machine_book_id"` // nullable
	SourceType    SourceType        `db:"source_type"`
	Status        ReservationStatus `db:"status"`
	ReservedAt    time.Time         `db:"reserved_at"`
	DueDate       time.Time         `db:"due_date"`
	ReturnedAt    *time.Time        `db:"returned_at"` // nullable
}
