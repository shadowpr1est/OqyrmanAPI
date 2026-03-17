package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type BookRepository interface {
	Create(ctx context.Context, book *entity.Book) (*entity.Book, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Book, int, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*entity.Book, error)
	ListByGenre(ctx context.Context, genreID uuid.UUID) ([]*entity.Book, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.Book, int, error)
	Update(ctx context.Context, book *entity.Book) (*entity.Book, error)
	UpdateCoverURL(ctx context.Context, id uuid.UUID, coverURL string) error // NEW
	Delete(ctx context.Context, id uuid.UUID) error
}
