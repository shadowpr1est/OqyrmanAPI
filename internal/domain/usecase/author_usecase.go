package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type AuthorUseCase interface {
	Create(ctx context.Context, author *entity.Author) (*entity.Author, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Author, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Author, int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.Author, int, error) // NEW
	Update(ctx context.Context, author *entity.Author) (*entity.Author, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
