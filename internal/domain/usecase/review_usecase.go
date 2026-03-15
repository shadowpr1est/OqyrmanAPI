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
	Update(ctx context.Context, review *entity.Review) (*entity.Review, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
