package genre

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type genreUseCase struct {
	genreRepo repository.GenreRepository
}

func NewGenreUseCase(genreRepo repository.GenreRepository) domainUseCase.GenreUseCase {
	return &genreUseCase{genreRepo: genreRepo}
}

func (u *genreUseCase) Create(ctx context.Context, genre *entity.Genre) (*entity.Genre, error) {
	genre.ID = uuid.New()
	return u.genreRepo.Create(ctx, genre)
}

func (u *genreUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Genre, error) {
	return u.genreRepo.GetByID(ctx, id)
}

func (u *genreUseCase) GetBySlug(ctx context.Context, slug string) (*entity.Genre, error) {
	return u.genreRepo.GetBySlug(ctx, slug)
}

func (u *genreUseCase) List(ctx context.Context) ([]*entity.Genre, error) {
	return u.genreRepo.List(ctx)
}

func (u *genreUseCase) Update(ctx context.Context, genre *entity.Genre) (*entity.Genre, error) {
	return u.genreRepo.Update(ctx, genre)
}

func (u *genreUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.genreRepo.Delete(ctx, id)
}
