package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type tokenRepo struct {
	db *sqlx.DB
}

func NewTokenRepo(db *sqlx.DB) *tokenRepo {
	return &tokenRepo{db: db}
}

func (r *tokenRepo) Save(ctx context.Context, token *entity.Token) error {
	query := `
		INSERT INTO tokens (id, user_id, refresh_token, expires_at, created_at)
		VALUES (:id, :user_id, :refresh_token, :expires_at, :created_at)`
	if _, err := r.db.NamedExecContext(ctx, query, token); err != nil {
		return fmt.Errorf("tokenRepo.Save: %w", err)
	}
	return nil
}

func (r *tokenRepo) GetByRefreshToken(ctx context.Context, refreshToken string) (*entity.Token, error) {
	var token entity.Token
	err := r.db.GetContext(ctx, &token, `SELECT * FROM tokens WHERE refresh_token = $1`, refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("tokenRepo.GetByRefreshToken: %w", err)
	}
	return &token, nil
}

func (r *tokenRepo) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM tokens WHERE refresh_token = $1`, refreshToken); err != nil {
		return fmt.Errorf("tokenRepo.DeleteByRefreshToken: %w", err)
	}
	return nil
}

func (r *tokenRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM tokens WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("tokenRepo.DeleteAllByUserID: %w", err)
	}
	return nil
}
