package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type GenreRepository interface {
	Create(ctx context.Context, genre *entity.Genre) (*entity.Genre, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Genre, error)
	List(ctx context.Context) ([]*entity.Genre, error)
	Update(ctx context.Context, genre *entity.Genre) (*entity.Genre, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
