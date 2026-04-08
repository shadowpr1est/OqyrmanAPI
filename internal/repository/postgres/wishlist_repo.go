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

type wishlistRepo struct {
	db *sqlx.DB
}

func NewWishlistRepo(db *sqlx.DB) *wishlistRepo {
	return &wishlistRepo{db: db}
}

func (r *wishlistRepo) Add(ctx context.Context, userID, bookID uuid.UUID, status entity.ShelfStatus) (*entity.Wishlist, error) {
	var w entity.Wishlist
	err := r.db.GetContext(ctx, &w, `
		INSERT INTO wishlists (id, user_id, book_id, status, added_at)
		VALUES (gen_random_uuid(), $1, $2, $3, now())
		ON CONFLICT (user_id, book_id)
		DO UPDATE SET status = $3
		RETURNING *`,
		userID, bookID, status,
	)
	if err != nil {
		return nil, fmt.Errorf("wishlistRepo.Add: %w", err)
	}
	return &w, nil
}

func (r *wishlistRepo) Remove(ctx context.Context, userID, bookID uuid.UUID) error {
	query := `DELETE FROM wishlists WHERE user_id = $1 AND book_id = $2`
	if _, err := r.db.ExecContext(ctx, query, userID, bookID); err != nil {
		return fmt.Errorf("wishlistRepo.Remove: %w", err)
	}
	return nil
}

func (r *wishlistRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Wishlist, error) {
	var items []*entity.Wishlist
	query := `SELECT * FROM wishlists WHERE user_id = $1 ORDER BY added_at DESC`
	if err := r.db.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, fmt.Errorf("wishlistRepo.ListByUser: %w", err)
	}
	return items, nil
}

func (r *wishlistRepo) ListByUserView(ctx context.Context, userID uuid.UUID, status *entity.ShelfStatus) ([]*entity.WishlistView, error) {
	var items []*entity.WishlistView

	query := `
		SELECT w.id, w.user_id, w.book_id,
		       b.title AS book_title, COALESCE(b.cover_url, '') AS book_cover_url,
		       b.avg_rating AS book_avg_rating,
		       b.author_id, a.name AS author_name,
		       b.genre_id, g.name AS genre_name,
		       w.status,
		       w.added_at
		FROM wishlists w
		JOIN books   b ON b.id = w.book_id    AND b.deleted_at IS NULL
		JOIN authors a ON a.id = b.author_id  AND a.deleted_at IS NULL
		JOIN genres  g ON g.id = b.genre_id   AND g.deleted_at IS NULL
		WHERE w.user_id = $1`

	args := []interface{}{userID}

	if status != nil {
		query += ` AND w.status = $2`
		args = append(args, *status)
	}

	query += ` ORDER BY w.added_at DESC`

	if err := r.db.SelectContext(ctx, &items, query, args...); err != nil {
		return nil, fmt.Errorf("wishlistRepo.ListByUserView: %w", err)
	}
	return items, nil
}

func (r *wishlistRepo) ExistsByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND book_id = $2)`
	if err := r.db.GetContext(ctx, &exists, query, userID, bookID); err != nil {
		return false, fmt.Errorf("wishlistRepo.ExistsByUserAndBook: %w", err)
	}
	return exists, nil
}

func (r *wishlistRepo) UpdateStatus(ctx context.Context, userID, bookID uuid.UUID, status entity.ShelfStatus) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE wishlists SET status = $3 WHERE user_id = $1 AND book_id = $2`,
		userID, bookID, status,
	)
	if err != nil {
		return fmt.Errorf("wishlistRepo.UpdateStatus: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *wishlistRepo) GetStatusByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*entity.ShelfStatus, error) {
	var status entity.ShelfStatus
	err := r.db.GetContext(ctx, &status,
		`SELECT status FROM wishlists WHERE user_id = $1 AND book_id = $2`,
		userID, bookID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("wishlistRepo.GetStatusByUserAndBook: %w", err)
	}
	return &status, nil
}