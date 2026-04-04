package usecase

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type AuthorUseCase interface {
	Create(ctx context.Context, author *entity.Author) (*entity.Author, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Author, error)
	List(ctx context.Context, limit, offset int) ([]*entity.Author, int, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entity.Author, int, error)
	Update(ctx context.Context, author *entity.Author) (*entity.Author, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UploadPhoto(ctx context.Context, id uuid.UUID, file io.Reader, size int64, contentType string) (*entity.Author, error)
}
