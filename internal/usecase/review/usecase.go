package review

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type reviewUseCase struct {
	reviewRepo repository.ReviewRepository
}

func NewReviewUseCase(
	reviewRepo repository.ReviewRepository,
	bookRepo repository.BookRepository,
) domainUseCase.ReviewUseCase {
	return &reviewUseCase{
		reviewRepo: reviewRepo,
	}
}

func (u *reviewUseCase) Create(ctx context.Context, review *entity.Review) (*entity.Review, error) {
	if review.Rating < 1 || review.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	review.ID = uuid.New()
	review.CreatedAt = time.Now()

	return u.reviewRepo.Create(ctx, review)
}

func (u *reviewUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Review, error) {
	return u.reviewRepo.GetByID(ctx, id)
}

func (u *reviewUseCase) ListByBook(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.Review, int, error) {
	return u.reviewRepo.ListByBook(ctx, bookID, limit, offset)
}

func (u *reviewUseCase) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Review, error) {
	return u.reviewRepo.ListByUser(ctx, userID)
}

func (u *reviewUseCase) Update(ctx context.Context, review *entity.Review) (*entity.Review, error) {
	if review.Rating < 1 || review.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	return u.reviewRepo.Update(ctx, review)
}

func (u *reviewUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.reviewRepo.Delete(ctx, id)
}
