package postgres

import (
	"context"
	"fmt"

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
