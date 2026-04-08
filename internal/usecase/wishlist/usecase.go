package wishlist

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type wishlistUseCase struct {
	wishlistRepo repository.WishlistRepository
}

func NewWishlistUseCase(wishlistRepo repository.WishlistRepository) domainUseCase.WishlistUseCase {
	return &wishlistUseCase{wishlistRepo: wishlistRepo}
}

func (u *wishlistUseCase) Add(ctx context.Context, userID, bookID uuid.UUID, status entity.ShelfStatus) (*entity.Wishlist, error) {
	return u.wishlistRepo.Add(ctx, userID, bookID, status)
}

func (u *wishlistUseCase) Remove(ctx context.Context, userID, bookID uuid.UUID) error {
	return u.wishlistRepo.Remove(ctx, userID, bookID)
}

func (u *wishlistUseCase) UpdateStatus(ctx context.Context, userID, bookID uuid.UUID, status entity.ShelfStatus) error {
	return u.wishlistRepo.UpdateStatus(ctx, userID, bookID, status)
}

func (u *wishlistUseCase) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Wishlist, error) {
	return u.wishlistRepo.ListByUser(ctx, userID)
}

func (u *wishlistUseCase) ExistsByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (bool, error) {
	return u.wishlistRepo.ExistsByUserAndBook(ctx, userID, bookID)
}

func (u *wishlistUseCase) GetStatusByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*entity.ShelfStatus, error) {
	return u.wishlistRepo.GetStatusByUserAndBook(ctx, userID, bookID)
}

func (u *wishlistUseCase) ListByUserView(ctx context.Context, userID uuid.UUID, status *entity.ShelfStatus) ([]*entity.WishlistView, error) {
	return u.wishlistRepo.ListByUserView(ctx, userID, status)
}