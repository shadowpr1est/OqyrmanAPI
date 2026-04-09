package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
)

type loginAttemptRepo struct {
	db *sqlx.DB
}

func NewLoginAttemptRepo(db *sqlx.DB) repository.LoginAttemptRepository {
	return &loginAttemptRepo{db: db}
}

func (r *loginAttemptRepo) RecordFailedAttempt(ctx context.Context, email string, maxAttempts int, lockoutDuration time.Duration) (locked bool, err error) {
	// Single atomic query: upsert the attempt, reset if lockout has expired,
	// and set locked_at when count reaches maxAttempts.
	lockoutSeconds := lockoutDuration.Seconds()

	var count int
	var lockedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, `
		INSERT INTO login_attempts (email, count, last_attempt_at)
		VALUES ($1, 1, now())
		ON CONFLICT (email) DO UPDATE
		    SET count = CASE
		            -- lockout expired: reset counter
		            WHEN login_attempts.locked_at IS NOT NULL
		                 AND login_attempts.locked_at + $3::interval < now()
		            THEN 1
		            ELSE login_attempts.count + 1
		        END,
		        locked_at = CASE
		            -- lockout expired: clear lock, let new counter start
		            WHEN login_attempts.locked_at IS NOT NULL
		                 AND login_attempts.locked_at + $3::interval < now()
		            THEN NULL
		            -- stale record (no recent attempts): reset
		            WHEN login_attempts.locked_at IS NULL
		                 AND login_attempts.last_attempt_at + $3::interval < now()
		            THEN NULL
		            -- reached max attempts: lock now
		            WHEN login_attempts.count + 1 >= $2
		            THEN now()
		            ELSE login_attempts.locked_at
		        END,
		        last_attempt_at = now()
		RETURNING count, locked_at`,
		email, maxAttempts, fmt.Sprintf("%d seconds", int(lockoutSeconds)),
	).Scan(&count, &lockedAt)
	if err != nil {
		return false, fmt.Errorf("loginAttemptRepo.RecordFailedAttempt: %w", err)
	}

	return lockedAt.Valid, nil
}

func (r *loginAttemptRepo) IsLocked(ctx context.Context, email string, lockoutDuration time.Duration) (bool, error) {
	var lockedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT locked_at FROM login_attempts
		WHERE email = $1`, email,
	).Scan(&lockedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("loginAttemptRepo.IsLocked: %w", err)
	}

	if !lockedAt.Valid {
		return false, nil
	}

	// Lock still active?
	if time.Since(lockedAt.Time) < lockoutDuration {
		return true, nil
	}

	// Lock expired — clean up stale record in background.
	go func() {
		_, delErr := r.db.ExecContext(context.Background(),
			`DELETE FROM login_attempts WHERE email = $1 AND locked_at IS NOT NULL AND locked_at + $2::interval < now()`,
			email, fmt.Sprintf("%d seconds", int(lockoutDuration.Seconds())),
		)
		if delErr != nil {
			slog.Warn("loginAttemptRepo: failed to clean expired lock", "err", delErr, "email", email)
		}
	}()

	return false, nil
}

func (r *loginAttemptRepo) Reset(ctx context.Context, email string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM login_attempts WHERE email = $1`, email,
	)
	if err != nil {
		return fmt.Errorf("loginAttemptRepo.Reset: %w", err)
	}
	return nil
}
