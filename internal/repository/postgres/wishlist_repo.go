package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type wishlistRepo struct {
	db *sqlx.DB
}

func NewWishlistRepo(db *sqlx.DB) *wishlistRepo {
	return &wishlistRepo{db: db}
}

func (r *wishlistRepo) Add(ctx context.Context, userID, bookID uuid.UUID) (*entity.Wishlist, error) {
	var w entity.Wishlist
	err := r.db.GetContext(ctx, &w, `
		INSERT INTO wishlists (id, user_id, book_id, added_at)
		VALUES (gen_random_uuid(), $1, $2, now())
		RETURNING *`,
		userID, bookID,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, entity.ErrDuplicateWishlist
		}
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

func (r *wishlistRepo) ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.WishlistView, error) {
	var items []*entity.WishlistView
	if err := r.db.SelectContext(ctx, &items, `
		SELECT w.id, w.user_id, w.book_id,
		       b.title AS book_title, COALESCE(b.cover_url, '') AS book_cover_url,
		       b.avg_rating AS book_avg_rating,
		       b.author_id, a.name AS author_name,
		       b.genre_id, g.name AS genre_name,
		       w.added_at
		FROM wishlists w
		JOIN books   b ON b.id = w.book_id    AND b.deleted_at IS NULL
		JOIN authors a ON a.id = b.author_id  AND a.deleted_at IS NULL
		JOIN genres  g ON g.id = b.genre_id   AND g.deleted_at IS NULL
		WHERE w.user_id = $1
		ORDER BY w.added_at DESC`,
		userID,
	); err != nil {
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
