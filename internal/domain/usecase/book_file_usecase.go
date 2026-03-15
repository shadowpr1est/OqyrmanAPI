package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type BookFileUseCase interface {
	Create(ctx context.Context, file *entity.BookFile) (*entity.BookFile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.BookFile, error)
	ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookFile, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
