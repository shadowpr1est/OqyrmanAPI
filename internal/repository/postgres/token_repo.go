package postgres

import (
	"context"
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
	query := `SELECT * FROM tokens WHERE refresh_token = $1`
	if err := r.db.GetContext(ctx, &token, query, refreshToken); err != nil {
		return nil, fmt.Errorf("tokenRepo.GetByRefreshToken: %w", err)
	}
	return &token, nil
}

func (r *tokenRepo) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	query := `DELETE FROM tokens WHERE refresh_token = $1`
	if _, err := r.db.ExecContext(ctx, query, refreshToken); err != nil {
		return fmt.Errorf("tokenRepo.DeleteByRefreshToken: %w", err)
	}
	return nil
}

func (r *tokenRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM tokens WHERE user_id = $1`
	if _, err := r.db.ExecContext(ctx, query, userID); err != nil {
		return fmt.Errorf("tokenRepo.DeleteAllByUserID: %w", err)
	}
	return nil
}
