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
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("reviewRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("reviewRepo.Create: no rows returned")
	}
	if err := rows.StructScan(review); err != nil {
		return nil, fmt.Errorf("reviewRepo.Create scan: %w", err)
	}
	return review, nil
}

func (r *reviewRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Review, error) {
	var review entity.Review
	err := r.db.GetContext(ctx, &review,
		`SELECT * FROM reviews WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("reviewRepo.GetByID: %w", err)
	}
	return &review, nil
}

func (r *reviewRepo) ListByBook(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.Review, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM reviews
		WHERE book_id = $1 AND deleted_at IS NULL`,
		bookID,
	); err != nil {
		return nil, 0, fmt.Errorf("reviewRepo.ListByBook count: %w", err)
	}

	var reviews []*entity.Review
	if err := r.db.SelectContext(ctx, &reviews, `
		SELECT * FROM reviews
		WHERE book_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		bookID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reviewRepo.ListByBook: %w", err)
	}

	return reviews, total, nil
}

func (r *reviewRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Review, error) {
	var reviews []*entity.Review
	if err := r.db.SelectContext(ctx, &reviews, `
		SELECT * FROM reviews
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`,
		userID,
	); err != nil {
		return nil, fmt.Errorf("reviewRepo.ListByUser: %w", err)
	}
	return reviews, nil
}

func (r *reviewRepo) Update(ctx context.Context, review *entity.Review) (*entity.Review, error) {
	query := `
		UPDATE reviews
		SET rating = :rating, body = :body
		WHERE id = :id AND deleted_at IS NULL
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, review)
	if err != nil {
		return nil, fmt.Errorf("reviewRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("reviewRepo.Update rows error: %w", err)
		}
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(review); err != nil {
		return nil, fmt.Errorf("reviewRepo.Update scan: %w", err)
	}
	return review, nil
}

func (r *reviewRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE reviews SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("reviewRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reviewRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}
