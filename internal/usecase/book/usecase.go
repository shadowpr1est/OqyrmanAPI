package book

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type bookUseCase struct {
	bookRepo repository.BookRepository
}

func NewBookUseCase(bookRepo repository.BookRepository) domainUseCase.BookUseCase {
	return &bookUseCase{bookRepo: bookRepo}
}

func (u *bookUseCase) Create(ctx context.Context, book *entity.Book) (*entity.Book, error) {
	book.ID = uuid.New()
	return u.bookRepo.Create(ctx, book)
}

func (u *bookUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error) {
	return u.bookRepo.GetByID(ctx, id)
}

func (u *bookUseCase) List(ctx context.Context, limit, offset int) ([]*entity.Book, int, error) {
	return u.bookRepo.List(ctx, limit, offset)
}

func (u *bookUseCase) ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*entity.Book, error) {
	return u.bookRepo.ListByAuthor(ctx, authorID)
}

func (u *bookUseCase) ListByGenre(ctx context.Context, genreID uuid.UUID) ([]*entity.Book, error) {
	return u.bookRepo.ListByGenre(ctx, genreID)
}

func (u *bookUseCase) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Book, int, error) {
	return u.bookRepo.Search(ctx, query, limit, offset)
}

func (u *bookUseCase) Update(ctx context.Context, book *entity.Book) (*entity.Book, error) {
	return u.bookRepo.Update(ctx, book)
}

func (u *bookUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.bookRepo.Delete(ctx, id)
}
