package repository

import (
	"context"
	"time"
)

type LoginAttemptRepository interface {
	RecordFailedAttempt(ctx context.Context, email string, maxAttempts int, lockoutDuration time.Duration) (locked bool, err error)
	IsLocked(ctx context.Context, email string, lockoutDuration time.Duration) (bool, error)
	Reset(ctx context.Context, email string) error
}
