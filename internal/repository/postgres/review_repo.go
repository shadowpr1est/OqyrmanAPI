package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type reviewRepo struct {
	db *sqlx.DB
}

func NewReviewRepo(db *sqlx.DB) *reviewRepo {
	return &reviewRepo{db: db}
}

func (r *reviewRepo) Create(ctx context.Context, review *entity.Review) (*entity.Review, error) {
	query := `
		INSERT INTO reviews (id, user_id, book_id, rating, body, created_at)
		VALUES (:id, :user_id, :book_id, :rating, :body, :created_at)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, review)
	if err != nil {
		return nil, fmt.Errorf("reviewRepo.Create: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(review); err != nil {
		return nil, fmt.Errorf("reviewRepo.Create scan: %w", err)
	}
	return review, nil
}

func (r *reviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Review, error) {
	var review entity.Review
	query := `SELECT * FROM reviews WHERE id = $1`
	if err := r.db.GetContext(ctx, &review, query, id); err != nil {
		return nil, fmt.Errorf("reviewRepo.GetByID: %w", err)
	}
	return &review, nil
}

func (r *reviewRepo) ListByBook(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.Review, int, error) {
	var reviews []*entity.Review
	var total int
	query := `SELECT * FROM reviews WHERE book_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &reviews, query, bookID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("reviewRepo.ListByBook: %w", err)
	}
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM reviews WHERE book_id = $1`, bookID); err != nil {
		return nil, 0, fmt.Errorf("reviewRepo.ListByBook count: %w", err)
	}
	return reviews, total, nil
}

func (r *reviewRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Review, error) {
	var reviews []*entity.Review
	query := `SELECT * FROM reviews WHERE user_id = $1 ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &reviews, query, userID); err != nil {
		return nil, fmt.Errorf("reviewRepo.ListByUser: %w", err)
	}
	return reviews, nil
}

func (r *reviewRepo) Update(ctx context.Context, review *entity.Review) (*entity.Review, error) {
	query := `
		UPDATE reviews
		SET rating = :rating, body = :body
		WHERE id = :id
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, review)
	if err != nil {
		return nil, fmt.Errorf("reviewRepo.Update: %w", err)
	}
	defer rows.Close()
	rows.Next()
	if err := rows.StructScan(review); err != nil {
		return nil, fmt.Errorf("reviewRepo.Update scan: %w", err)
	}
	return review, nil
}

func (r *reviewRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reviews WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("reviewRepo.Delete: %w", err)
	}
	return nil
}
