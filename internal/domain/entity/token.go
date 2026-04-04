package entity

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID           uuid.UUID `db:"id"`
	UserID       uuid.UUID `db:"user_id"`
	RefreshToken string    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
	UserAgent    string    `db:"user_agent"`
	IP           string    `db:"ip"`
	CreatedAt    time.Time `db:"created_at"`
}
