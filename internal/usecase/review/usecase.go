package review

import (
	"context"
	"fmt"
	"log/slog"
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
	if reviewRepo == nil {
		panic("reviewUseCase: reviewRepo must not be nil")
	}
	if bookRepo == nil {
		panic("reviewUseCase: bookRepo must not be nil")
	}
	return &reviewUseCase{
		reviewRepo: reviewRepo,
		bookRepo:   bookRepo,
	}
}

func (u *reviewUseCase) Create(ctx context.Context, review *entity.Review) (*entity.Review, error) {
	if review.Rating < 1 || review.Rating > 5 {
		return nil, fmt.Errorf("%w: rating must be between 1 and 5", entity.ErrValidation)
	}
	review.ID = uuid.New()
	review.CreatedAt = time.Now()

	created, err := u.reviewRepo.Create(ctx, review)
	if err != nil {
		return nil, err
	}
	if err := u.bookRepo.UpdateRating(ctx, created.BookID); err != nil {
		slog.WarnContext(ctx, "failed to update book rating after review create",
			"book_id", created.BookID, "err", err)
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

func (u *reviewUseCase) Update(ctx context.Context, review *entity.Review, callerID uuid.UUID) (*entity.Review, error) {
	if review.Rating < 1 || review.Rating > 5 {
		return nil, fmt.Errorf("%w: rating must be between 1 and 5", entity.ErrValidation)
	}
	if review.UserID != callerID {
		return nil, entity.ErrForbidden
	}

	updated, err := u.reviewRepo.Update(ctx, review)
	if err != nil {
		return nil, err
	}
	if err := u.bookRepo.UpdateRating(ctx, updated.BookID); err != nil {
		slog.WarnContext(ctx, "failed to update book rating after review update",
			"book_id", updated.BookID, "err", err)
	}
	return updated, nil
}

func (u *reviewUseCase) Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	existing, err := u.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.UserID != callerID {
		return entity.ErrForbidden
	}
	if err := u.reviewRepo.Delete(ctx, id); err != nil {
		return err
	}
	if err := u.bookRepo.UpdateRating(ctx, existing.BookID); err != nil {
		slog.WarnContext(ctx, "failed to update book rating after review delete",
			"book_id", existing.BookID, "err", err)
	}
	return nil
}

func (u *reviewUseCase) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReviewView, error) {
	return u.reviewRepo.GetByIDView(ctx, id)
}

func (u *reviewUseCase) ListByBookView(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.ReviewView, int, error) {
	return u.reviewRepo.ListByBookView(ctx, bookID, limit, offset)
}

func (u *reviewUseCase) ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReviewView, error) {
	return u.reviewRepo.ListByUserView(ctx, userID)
}
