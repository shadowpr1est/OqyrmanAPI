package postgres

import (
	"context"
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

func (r *wishlistRepo) Add(ctx context.Context, userID, bookID uuid.UUID) (*entity.Wishlist, error) {
	var w entity.Wishlist
	query := `
		INSERT INTO wishlists (id, user_id, book_id, added_at)
		VALUES (gen_random_uuid(), $1, $2, now())
		RETURNING *`
	if err := r.db.GetContext(ctx, &w, query, userID, bookID); err != nil {
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

func (r *wishlistRepo) ExistsByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM wishlists WHERE user_id = $1 AND book_id = $2)`
	if err := r.db.GetContext(ctx, &exists, query, userID, bookID); err != nil {
		return false, fmt.Errorf("wishlistRepo.ExistsByUserAndBook: %w", err)
	}
	return exists, nil
}
