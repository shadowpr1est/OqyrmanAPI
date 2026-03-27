package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
)

type BookUseCase interface {
	Create(ctx context.Context, book *entity.Book, cover *fileupload.File) (*entity.Book, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Book, int, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*entity.Book, error)
	ListByGenre(ctx context.Context, genreID uuid.UUID) ([]*entity.Book, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.Book, int, error)
	Update(ctx context.Context, book *entity.Book) (*entity.Book, error)
	UploadCover(ctx context.Context, id uuid.UUID, cover *fileupload.File) (*entity.Book, error)
	ListPopular(ctx context.Context, limit, offset int) ([]*entity.Book, int, error)
	ListSimilar(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.Book, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
