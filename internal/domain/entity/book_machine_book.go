package entity

import "github.com/google/uuid"

type BookMachineBook struct {
	ID              uuid.UUID `db:"id"`
	MachineID       uuid.UUID `db:"machine_id"`
	BookID          uuid.UUID `db:"book_id"`
	TotalCopies     int       `db:"total_copies"`
	AvailableCopies int       `db:"available_copies"`
}
