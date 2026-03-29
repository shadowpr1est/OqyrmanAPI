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
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reviewRepo.Create begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO reviews (id, user_id, book_id, rating, body, created_at)
		VALUES (:id, :user_id, :book_id, :rating, :body, :created_at)
		RETURNING *`
	rows, err := sqlx.NamedQueryContext(ctx, tx, query, review)
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
	rows.Close()

	if err := updateBookRatingTx(ctx, tx, review.BookID); err != nil {
		return nil, fmt.Errorf("reviewRepo.Create rating: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reviewRepo.Create commit: %w", err)
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
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reviewRepo.Update begin tx: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE reviews
		SET rating = :rating, body = :body
		WHERE id = :id AND deleted_at IS NULL
		RETURNING *`
	rows, err := sqlx.NamedQueryContext(ctx, tx, query, review)
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
	rows.Close()

	if err := updateBookRatingTx(ctx, tx, review.BookID); err != nil {
		return nil, fmt.Errorf("reviewRepo.Update rating: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reviewRepo.Update commit: %w", err)
	}
	return review, nil
}

func (r *reviewRepo) Delete(ctx context.Context, id uuid.UUID) error {
	// Получаем book_id до удаления — нужен для пересчёта рейтинга
	var bookID uuid.UUID
	if err := r.db.GetContext(ctx, &bookID,
		`SELECT book_id FROM reviews WHERE id = $1 AND deleted_at IS NULL`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrNotFound
		}
		return fmt.Errorf("reviewRepo.Delete get book_id: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("reviewRepo.Delete begin tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`UPDATE reviews SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("reviewRepo.Delete: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reviewRepo.Delete rows affected: %w", err)
	}
	if affected == 0 {
		return entity.ErrNotFound
	}

	if err := updateBookRatingTx(ctx, tx, bookID); err != nil {
		return fmt.Errorf("reviewRepo.Delete rating: %w", err)
	}

	return tx.Commit()
}

func (r *reviewRepo) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReviewView, error) {
	var v entity.ReviewView
	err := r.db.GetContext(ctx, &v, `
		SELECT rv.id, rv.user_id, u.full_name AS user_full_name,
		       COALESCE(u.avatar_url, '') AS user_avatar_url,
		       rv.book_id, b.title AS book_title,
		       rv.rating, rv.body, rv.created_at
		FROM reviews rv
		JOIN users u ON u.id = rv.user_id   AND u.deleted_at  IS NULL
		JOIN books b ON b.id = rv.book_id   AND b.deleted_at  IS NULL
		WHERE rv.id = $1 AND rv.deleted_at IS NULL`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("reviewRepo.GetByIDView: %w", err)
	}
	return &v, nil
}

func (r *reviewRepo) ListByBookView(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.ReviewView, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM reviews WHERE book_id = $1 AND deleted_at IS NULL`, bookID,
	); err != nil {
		return nil, 0, fmt.Errorf("reviewRepo.ListByBookView count: %w", err)
	}
	var items []*entity.ReviewView
	if err := r.db.SelectContext(ctx, &items, `
		SELECT rv.id, rv.user_id, u.full_name AS user_full_name,
		       COALESCE(u.avatar_url, '') AS user_avatar_url,
		       rv.book_id, b.title AS book_title,
		       rv.rating, rv.body, rv.created_at
		FROM reviews rv
		JOIN users u ON u.id = rv.user_id   AND u.deleted_at  IS NULL
		JOIN books b ON b.id = rv.book_id   AND b.deleted_at  IS NULL
		WHERE rv.book_id = $1 AND rv.deleted_at IS NULL
		ORDER BY rv.created_at DESC
		LIMIT $2 OFFSET $3`,
		bookID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reviewRepo.ListByBookView: %w", err)
	}
	return items, total, nil
}

func (r *reviewRepo) ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReviewView, error) {
	var items []*entity.ReviewView
	if err := r.db.SelectContext(ctx, &items, `
		SELECT rv.id, rv.user_id, u.full_name AS user_full_name,
		       COALESCE(u.avatar_url, '') AS user_avatar_url,
		       rv.book_id, b.title AS book_title,
		       rv.rating, rv.body, rv.created_at
		FROM reviews rv
		JOIN users u ON u.id = rv.user_id   AND u.deleted_at  IS NULL
		JOIN books b ON b.id = rv.book_id   AND b.deleted_at  IS NULL
		WHERE rv.user_id = $1 AND rv.deleted_at IS NULL
		ORDER BY rv.created_at DESC`,
		userID,
	); err != nil {
		return nil, fmt.Errorf("reviewRepo.ListByUserView: %w", err)
	}
	return items, nil
}

// updateBookRatingTx пересчитывает avg_rating книги внутри существующей транзакции.
func updateBookRatingTx(ctx context.Context, tx *sqlx.Tx, bookID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE books
		SET avg_rating = (
			SELECT COALESCE(AVG(rating), 0)
			FROM reviews
			WHERE book_id = $1 AND deleted_at IS NULL
		)
		WHERE id = $1 AND deleted_at IS NULL`, bookID)
	return err
}
