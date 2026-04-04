package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCodeNotFound         = errors.New("verification code not found or expired")
	ErrAlreadyVerified      = errors.New("email already verified")
	ErrEmailNotFound        = errors.New("user with this email not found")
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrRegistrationPending  = errors.New("registration pending: verification code is still active, please check your email or wait 5 minutes")
)

type EmailVerificationCode struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Code      string    `db:"code"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
