package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type WishlistUseCase interface {
	Add(ctx context.Context, userID, bookID uuid.UUID) (*entity.Wishlist, error)
	Remove(ctx context.Context, userID, bookID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Wishlist, error)
	ExistsByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error)

	// View method — returns enriched nested data for GET endpoint.
	ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.WishlistView, error)
}
