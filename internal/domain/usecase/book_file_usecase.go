package usecase

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type BookFileUseCase interface {
	Create(ctx context.Context, file *entity.BookFile) (*entity.BookFile, error)
	Upload(ctx context.Context, bookID uuid.UUID, format string, isAudio bool, filename string, reader io.Reader, size int64, contentType string) (*entity.BookFile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.BookFile, error)
	ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookFile, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
