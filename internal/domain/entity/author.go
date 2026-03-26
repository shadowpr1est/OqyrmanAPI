package entity

import (
	"time"

	"github.com/google/uuid"
)

type Author struct {
	ID        uuid.UUID  `db:"id"`
	Name      string     `db:"name"`
	Bio       string     `db:"bio"`
	BirthDate *time.Time `db:"birth_date"`
	DeathDate *time.Time `db:"death_date"`
	PhotoURL  string     `db:"photo_url"`
	CreatedAt time.Time  `db:"created_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}
