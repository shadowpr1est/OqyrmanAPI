package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
)

type BookFileUseCase interface {
	// Upload validates, stores the file in object storage, and creates the DB record.
	// isAudio is derived from format — never accepted from the caller.
	// totalPages is optional and only applied for document formats (PDF/EPUB).
	Upload(ctx context.Context, bookID uuid.UUID, format entity.BookFileFormat, totalPages *int, file *fileupload.File) (*entity.BookFile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.BookFile, error)
	ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookFile, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
