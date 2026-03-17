package author

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type authorUseCase struct {
	authorRepo repository.AuthorRepository
}

func NewAuthorUseCase(authorRepo repository.AuthorRepository) domainUseCase.AuthorUseCase {
	return &authorUseCase{authorRepo: authorRepo}
}

func (u *authorUseCase) Create(ctx context.Context, author *entity.Author) (*entity.Author, error) {
	author.ID = uuid.New()
	return u.authorRepo.Create(ctx, author)
}

func (u *authorUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Author, error) {
	return u.authorRepo.GetByID(ctx, id)
}

func (u *authorUseCase) List(ctx context.Context, limit, offset int) ([]*entity.Author, int, error) {
	return u.authorRepo.List(ctx, limit, offset)
}

func (u *authorUseCase) Update(ctx context.Context, author *entity.Author) (*entity.Author, error) {
	return u.authorRepo.Update(ctx, author)
}

func (u *authorUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.authorRepo.Delete(ctx, id)
}

func (u *authorUseCase) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Author, int, error) {
	return u.authorRepo.Search(ctx, query, limit, offset)
}
