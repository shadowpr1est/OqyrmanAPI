package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type statsRepo struct {
	db *sqlx.DB
}

func NewStatsRepo(db *sqlx.DB) *statsRepo {
	return &statsRepo{db: db}
}

func (r *statsRepo) GetStats(ctx context.Context) (*entity.Stats, error) {
	var stats entity.Stats
	query := `
		SELECT
			(SELECT COUNT(*) FROM users)                                 AS users_total,
			(SELECT COUNT(*) FROM books)                                 AS books_total,
			(SELECT COUNT(*) FROM authors)                               AS authors_total,
			(SELECT COUNT(*) FROM reservations WHERE status = 'active')  AS reservations_active,
			(SELECT COUNT(*) FROM reservations WHERE status = 'pending') AS reservations_pending,
			(SELECT COUNT(*) FROM reservations)                          AS reservations_total,
			(SELECT COUNT(*) FROM reviews)                               AS reviews_total`
	if err := r.db.GetContext(ctx, &stats, query); err != nil {
		return nil, fmt.Errorf("statsRepo.GetStats: %w", err)
	}
	return &stats, nil
}

func (r *statsRepo) GetUserStats(ctx context.Context, userID uuid.UUID) (*entity.UserStats, error) {
	var stats entity.UserStats
	query := `
		SELECT
			(SELECT COUNT(*) FROM reading_sessions
			 WHERE user_id = $1 AND status = 'finished') AS books_read,
			(SELECT COUNT(*) FROM reservations
			 WHERE user_id = $1 AND status = 'active') AS active_reservations,
			(SELECT COUNT(*) FROM reviews
			 WHERE user_id = $1 AND deleted_at IS NULL) AS reviews_given,
			(SELECT COUNT(*) FROM wishlists
			 WHERE user_id = $1) AS wishlist_count`
	if err := r.db.GetContext(ctx, &stats, query, userID); err != nil {
		return nil, fmt.Errorf("statsRepo.GetUserStats: %w", err)
	}
	return &stats, nil
}
