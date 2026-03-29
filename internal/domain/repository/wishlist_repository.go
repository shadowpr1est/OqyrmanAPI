package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type WishlistRepository interface {
	Add(ctx context.Context, userID, bookID uuid.UUID) (*entity.Wishlist, error)
	Remove(ctx context.Context, userID, bookID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Wishlist, error)
	ExistsByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error)

	// View method — used by GET /wishlist; returns joined book/author/genre data.
	ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.WishlistView, error)
}
