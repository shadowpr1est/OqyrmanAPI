package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *entity.Review) (*entity.Review, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Review, error)
	ListByBook(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.Review, int, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Review, error)
	Update(ctx context.Context, review *entity.Review) (*entity.Review, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// View methods — used by GET endpoints; return joined user/book data.
	GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReviewView, error)
	ListByBookView(ctx context.Context, bookID uuid.UUID, limit, offset int) ([]*entity.ReviewView, int, error)
	ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReviewView, error)
}
