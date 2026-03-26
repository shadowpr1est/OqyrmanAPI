package entity

import (
	"time"

	"github.com/google/uuid"
)

type Genre struct {
	ID        uuid.UUID  `db:"id"`
	Name      string     `db:"name"`
	Slug      string     `db:"slug"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}
