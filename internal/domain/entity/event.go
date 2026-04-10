package entity

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID           uuid.UUID  `db:"id"`
	Title        string     `db:"title"`
	Description  *string    `db:"description"`
	CoverURL     *string    `db:"cover_url"`
	Location     *string    `db:"location"`
	StartsAt     time.Time  `db:"starts_at"`
	EndsAt       time.Time  `db:"ends_at"`
	ReminderSent bool       `db:"reminder_sent"`
	CreatedAt    time.Time  `db:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}
