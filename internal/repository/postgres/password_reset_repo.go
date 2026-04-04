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

type passwordResetCodeRepo struct {
	db *sqlx.DB
}

func NewPasswordResetCodeRepo(db *sqlx.DB) repository.PasswordResetCodeRepository {
	return &passwordResetCodeRepo{db: db}
}

func (r *passwordResetCodeRepo) Save(ctx context.Context, code *entity.PasswordResetCode) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO password_reset_codes (id, user_id, code, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		    SET code       = EXCLUDED.code,
		        expires_at = EXCLUDED.expires_at,
		        created_at = EXCLUDED.created_at`,
		code.ID, code.UserID, code.Code, code.ExpiresAt, code.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("passwordResetCodeRepo.Save: %w", err)
	}
	return nil
}

func (r *passwordResetCodeRepo) GetByUserAndCode(ctx context.Context, userID uuid.UUID, code string) (*entity.PasswordResetCode, error) {
	var c entity.PasswordResetCode
	err := r.db.GetContext(ctx, &c, `
		SELECT id, user_id, code, expires_at, created_at
		FROM password_reset_codes
		WHERE user_id = $1 AND code = $2`,
		userID, code,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrResetCodeNotFound
		}
		return nil, fmt.Errorf("passwordResetCodeRepo.GetByUserAndCode: %w", err)
	}
	return &c, nil
}

func (r *passwordResetCodeRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM password_reset_codes WHERE user_id = $1`, userID,
	)
	return err
}
