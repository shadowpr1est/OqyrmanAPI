package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReviewUseCase interface {
	Create(ctx context.Context, review *entity.Review) (*entity.Review, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Review, error)
	ListByBook(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.Review, int, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Review, error)
	// Update validates ownership: returns ErrForbidden if callerID != review.UserID.
	Update(ctx context.Context, review *entity.Review, callerID uuid.UUID) (*entity.Review, error)
	// Delete validates ownership: returns ErrForbidden if callerID != review.UserID.
	Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error
}
