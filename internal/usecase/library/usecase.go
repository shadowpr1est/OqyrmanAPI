package library

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type libraryUseCase struct {
	libraryRepo repository.LibraryRepository
}

func NewLibraryUseCase(libraryRepo repository.LibraryRepository) domainUseCase.LibraryUseCase {
	return &libraryUseCase{libraryRepo: libraryRepo}
}

func (u *libraryUseCase) Create(ctx context.Context, library *entity.Library) (*entity.Library, error) {
	library.ID = uuid.New()
	return u.libraryRepo.Create(ctx, library)
}

func (u *libraryUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Library, error) {
	return u.libraryRepo.GetByID(ctx, id)
}

func (u *libraryUseCase) List(ctx context.Context, limit, offset int) ([]*entity.Library, int, error) {
	return u.libraryRepo.List(ctx, limit, offset)
}

func (u *libraryUseCase) ListNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Library, error) {
	return u.libraryRepo.ListNearby(ctx, lat, lng, radiusKm)
}

func (u *libraryUseCase) Update(ctx context.Context, library *entity.Library) (*entity.Library, error) {
	return u.libraryRepo.Update(ctx, library)
}

func (u *libraryUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.libraryRepo.Delete(ctx, id)
}
