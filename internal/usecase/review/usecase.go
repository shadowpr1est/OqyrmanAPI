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
	bookRepo   repository.BookRepository
}

func NewReviewUseCase(
	reviewRepo repository.ReviewRepository,
	bookRepo repository.BookRepository,
) domainUseCase.ReviewUseCase {
	return &reviewUseCase{
		reviewRepo: reviewRepo,
		bookRepo:   bookRepo,
	}
}

func (u *reviewUseCase) Create(ctx context.Context, review *entity.Review) (*entity.Review, error) {
	if review.Rating < 1 || review.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	review.ID = uuid.New()
	review.CreatedAt = time.Now()

	created, err := u.reviewRepo.Create(ctx, review)
	if err != nil {
		return nil, err
	}
	if u.bookRepo != nil {
		_ = u.bookRepo.UpdateRating(ctx, created.BookID)
	}
	return created, nil
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

	updated, err := u.reviewRepo.Update(ctx, review)
	if err != nil {
		return nil, err
	}
	if u.bookRepo != nil {
		_ = u.bookRepo.UpdateRating(ctx, updated.BookID)
	}
	return updated, nil
}

func (u *reviewUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := u.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.reviewRepo.Delete(ctx, id); err != nil {
		return err
	}
	if u.bookRepo != nil {
		_ = u.bookRepo.UpdateRating(ctx, existing.BookID)
	}
	return nil
}
