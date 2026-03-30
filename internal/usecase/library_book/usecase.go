package library_book

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type libraryBookUseCase struct {
	libraryBookRepo repository.LibraryBookRepository
}

func NewLibraryBookUseCase(libraryBookRepo repository.LibraryBookRepository) domainUseCase.LibraryBookUseCase {
	return &libraryBookUseCase{libraryBookRepo: libraryBookRepo}
}

func (u *libraryBookUseCase) Create(ctx context.Context, lb *entity.LibraryBook) (*entity.LibraryBook, error) {
	lb.ID = uuid.New()
	return u.libraryBookRepo.Create(ctx, lb)
}

func (u *libraryBookUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.LibraryBook, error) {
	return u.libraryBookRepo.GetByID(ctx, id)
}

func (u *libraryBookUseCase) ListByLibrary(ctx context.Context, libraryID uuid.UUID) ([]*entity.LibraryBook, error) {
	return u.libraryBookRepo.ListByLibrary(ctx, libraryID)
}

func (u *libraryBookUseCase) ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.LibraryBook, error) {
	return u.libraryBookRepo.ListByBook(ctx, bookID)
}

func (u *libraryBookUseCase) Update(ctx context.Context, lb *entity.LibraryBook) (*entity.LibraryBook, error) {
	return u.libraryBookRepo.Update(ctx, lb)
}

func (u *libraryBookUseCase) UpdateCopies(ctx context.Context, id uuid.UUID, totalCopies *int, availableCopies *int) (*entity.LibraryBook, error) {
	return u.libraryBookRepo.UpdateCopies(ctx, id, totalCopies, availableCopies)
}

func (u *libraryBookUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.libraryBookRepo.Delete(ctx, id)
}

func (u *libraryBookUseCase) SearchInLibrary(ctx context.Context, libraryID uuid.UUID, q string, genreID *uuid.UUID, onlyAvailable bool, limit, offset int) ([]*entity.LibraryBookSearchResult, int, error) {
	return u.libraryBookRepo.SearchInLibrary(ctx, libraryID, q, genreID, onlyAvailable, limit, offset)
}

func (u *libraryBookUseCase) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.LibraryBookView, error) {
	return u.libraryBookRepo.GetByIDView(ctx, id)
}

func (u *libraryBookUseCase) ListByLibraryView(ctx context.Context, libraryID uuid.UUID) ([]*entity.LibraryBookView, error) {
	return u.libraryBookRepo.ListByLibraryView(ctx, libraryID)
}

func (u *libraryBookUseCase) ListByBookView(ctx context.Context, bookID uuid.UUID) ([]*entity.LibraryBookView, error) {
	return u.libraryBookRepo.ListByBookView(ctx, bookID)
}
