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
	UpdateCoverURL(ctx context.Context, id uuid.UUID, coverURL string) error
	UpdateRating(ctx context.Context, bookID uuid.UUID) error
	ListPopular(ctx context.Context, limit, offset int) ([]*entity.Book, int, error)
	ListSimilar(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.Book, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateTotalPages(ctx context.Context, bookID uuid.UUID, totalPages int) error

	// View methods — used by GET endpoints; return joined author/genre data.
	GetByIDView(ctx context.Context, id uuid.UUID) (*entity.BookView, error)
	ListView(ctx context.Context, limit, offset int) ([]*entity.BookView, int, error)
	ListByAuthorView(ctx context.Context, authorID uuid.UUID) ([]*entity.BookView, error)
	ListByGenreView(ctx context.Context, genreID uuid.UUID) ([]*entity.BookView, error)
	SearchView(ctx context.Context, query string, limit, offset int) ([]*entity.BookView, int, error)
	ListPopularView(ctx context.Context, limit, offset int) ([]*entity.BookView, int, error)
	ListSimilarView(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.BookView, error)
	ListRecommendedView(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.BookView, error)
}
