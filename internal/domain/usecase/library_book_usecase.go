package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type LibraryBookUseCase interface {
	Create(ctx context.Context, lb *entity.LibraryBook) (*entity.LibraryBook, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.LibraryBook, error)
	ListByLibrary(ctx context.Context, libraryID uuid.UUID) ([]*entity.LibraryBook, error)
	ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.LibraryBook, error)
	Update(ctx context.Context, lb *entity.LibraryBook) (*entity.LibraryBook, error)
	UpdateCopies(ctx context.Context, id uuid.UUID, totalCopies *int, availableCopies *int) (*entity.LibraryBook, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SearchInLibrary(ctx context.Context, libraryID uuid.UUID, q string, genreID *uuid.UUID, onlyAvailable bool, limit, offset int) ([]*entity.LibraryBookSearchResult, int, error)

	// View methods — return enriched nested data for GET endpoints.
	GetByIDView(ctx context.Context, id uuid.UUID) (*entity.LibraryBookView, error)
	ListByLibraryView(ctx context.Context, libraryID uuid.UUID) ([]*entity.LibraryBookView, error)
	ListByBookView(ctx context.Context, bookID uuid.UUID) ([]*entity.LibraryBookView, error)
}
