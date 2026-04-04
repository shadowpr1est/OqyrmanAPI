package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrResetCodeNotFound = errors.New("reset code not found or expired")

type PasswordResetCode struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Code      string    `db:"code"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
