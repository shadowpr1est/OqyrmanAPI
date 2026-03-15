package entity

import "github.com/google/uuid"

type Genre struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
	Slug string    `db:"slug"`
}
