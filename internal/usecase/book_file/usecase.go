package book_file

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type bookFileUseCase struct {
	bookFileRepo repository.BookFileRepository
}

func NewBookFileUseCase(bookFileRepo repository.BookFileRepository) domainUseCase.BookFileUseCase {
	return &bookFileUseCase{bookFileRepo: bookFileRepo}
}

func (u *bookFileUseCase) Create(ctx context.Context, file *entity.BookFile) (*entity.BookFile, error) {
	file.ID = uuid.New()
	return u.bookFileRepo.Create(ctx, file)
}

func (u *bookFileUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.BookFile, error) {
	return u.bookFileRepo.GetByID(ctx, id)
}

func (u *bookFileUseCase) ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookFile, error) {
	return u.bookFileRepo.ListByBook(ctx, bookID)
}

func (u *bookFileUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.bookFileRepo.Delete(ctx, id)
}
