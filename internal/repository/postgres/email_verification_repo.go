package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
)

type emailVerificationCodeRepo struct {
	db *sqlx.DB
}

func NewEmailVerificationCodeRepo(db *sqlx.DB) repository.EmailVerificationCodeRepository {
	return &emailVerificationCodeRepo{db: db}
}

func (r *emailVerificationCodeRepo) Save(ctx context.Context, code *entity.EmailVerificationCode) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO email_verification_codes (id, user_id, code, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		    SET code       = EXCLUDED.code,
		        expires_at = EXCLUDED.expires_at,
		        created_at = EXCLUDED.created_at`,
		code.ID, code.UserID, code.Code, code.ExpiresAt, code.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("emailVerificationCodeRepo.Save: %w", err)
	}
	return nil
}

func (r *emailVerificationCodeRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.EmailVerificationCode, error) {
	var c entity.EmailVerificationCode
	err := r.db.GetContext(ctx, &c, `
		SELECT id, user_id, code, expires_at, created_at
		FROM email_verification_codes
		WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrCodeNotFound
		}
		return nil, fmt.Errorf("emailVerificationCodeRepo.GetByUserID: %w", err)
	}
	return &c, nil
}

func (r *emailVerificationCodeRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM email_verification_codes WHERE user_id = $1`, userID,
	)
	return err
}
