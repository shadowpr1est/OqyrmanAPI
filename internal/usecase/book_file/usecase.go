package book_file

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainStorage "github.com/shadowpr1est/OqyrmanAPI/internal/domain/storage"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
)

type bookFileUseCase struct {
	bookFileRepo repository.BookFileRepository
	bookRepo     repository.BookRepository
	storage      domainStorage.FileStorage
}

func NewBookFileUseCase(
	bookFileRepo repository.BookFileRepository,
	bookRepo repository.BookRepository,
	storage domainStorage.FileStorage,
) domainUseCase.BookFileUseCase {
	return &bookFileUseCase{
		bookFileRepo: bookFileRepo,
		bookRepo:     bookRepo,
		storage:      storage,
	}
}

func (u *bookFileUseCase) Upload(
	ctx context.Context,
	bookID uuid.UUID,
	format entity.BookFileFormat,
	totalPages *int,
	file *fileupload.File,
) (*entity.BookFile, error) {
	if u.storage == nil {
		return nil, fmt.Errorf("file storage is not configured")
	}

	// Enforce file size limit before reading the full stream.
	maxSize := format.MaxSize()
	if file.Size > maxSize {
		return nil, fmt.Errorf("%w: file exceeds maximum allowed size of %d bytes", entity.ErrValidation, maxSize)
	}

	// Read first 512 bytes for magic bytes detection, then seek back to start.
	// This prevents uploading renamed files (e.g. shell.php renamed to book.pdf).
	buf := make([]byte, 512)
	n, err := file.Reader.Read(buf)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("bookFileUseCase.Upload: cannot read file header: %w", err)
	}
	if err := format.ValidateMagicBytes(buf[:n]); err != nil {
		return nil, err
	}
	if _, err := file.Reader.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("bookFileUseCase.Upload: cannot seek file: %w", err)
	}

	// isAudio is derived from format — never accepted from the caller.
	isAudio := format.IsAudio()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	objectKey := fmt.Sprintf("books/%s/%s%s", bookID.String(), uuid.New().String(), ext)

	fileURL, err := u.storage.Upload(ctx, objectKey, file.Reader, file.Size, file.ContentType)
	if err != nil {
		return nil, err
	}

	bookFile := &entity.BookFile{
		ID:      uuid.New(),
		BookID:  bookID,
		Format:  format,
		FileURL: fileURL,
		IsAudio: isAudio,
	}
	// DB-level UNIQUE (book_id, is_audio) constraint enforces the 1-audio / 1-document limit.
	// bookFileRepo.Create maps the pq 23505 error to entity.ErrFileLimitExceeded.
	result, err := u.bookFileRepo.Create(ctx, bookFile)
	if err != nil {
		return nil, err
	}

	// Update total_pages on the book for document uploads (non-fatal if it fails).
	if !isAudio && totalPages != nil && *totalPages > 0 {
		if err := u.bookRepo.UpdateTotalPages(ctx, bookID, *totalPages); err != nil {
			slog.WarnContext(ctx, "failed to update total_pages", "book_id", bookID, "err", err)
		}
	}

	return result, nil
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
