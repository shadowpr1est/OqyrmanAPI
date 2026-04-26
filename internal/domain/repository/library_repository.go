package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type LibraryRepository interface {
	Create(ctx context.Context, library *entity.Library) (*entity.Library, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Library, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Library, int, error)
	ListNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Library, error)
	Update(ctx context.Context, library *entity.Library) (*entity.Library, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePhotoURL(ctx context.Context, id uuid.UUID, url string) error
}
