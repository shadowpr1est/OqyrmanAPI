package book_file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/ledongthuc/pdf"
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
	file *fileupload.File,
) (*entity.BookFile, error) {
	if u.storage == nil {
		return nil, fmt.Errorf("file storage is not configured")
	}

	// Detect format from file extension.
	ext := strings.ToLower(filepath.Ext(file.Filename))
	var format entity.BookFileFormat
	switch ext {
	case ".pdf":
		format = entity.BookFileFormatPDF
	case ".epub":
		format = entity.BookFileFormatEPUB
	case ".mp3":
		format = entity.BookFileFormatMP3
	default:
		return nil, fmt.Errorf("%w: unsupported file extension %q, allowed: .pdf, .epub, .mp3", entity.ErrValidation, ext)
	}

	// Enforce file size limit before reading the full stream.
	if file.Size > format.MaxSize() {
		return nil, fmt.Errorf("%w: file exceeds maximum allowed size of %d bytes", entity.ErrValidation, format.MaxSize())
	}

	// Read entire file into memory for magic-byte validation, PDF page counting, and storage upload.
	data, err := io.ReadAll(file.Reader)
	if err != nil {
		return nil, fmt.Errorf("bookFileUseCase.Upload: cannot read file: %w", err)
	}

	// Validate magic bytes against declared format.
	headerLen := 512
	if len(data) < headerLen {
		headerLen = len(data)
	}
	if err := format.ValidateMagicBytes(data[:headerLen]); err != nil {
		return nil, err
	}

	// Detect content type from actual bytes for storage metadata.
	// For EPUB (ZIP), override to the correct MIME type.
	contentType := detectContentType(format, data[:headerLen])

	objectKey := fmt.Sprintf("books/%s/%s%s", bookID.String(), uuid.New().String(), ext)
	fileURL, err := u.storage.Upload(ctx, objectKey, bytes.NewReader(data), int64(len(data)), contentType)
	if err != nil {
		return nil, err
	}

	bookFile := &entity.BookFile{
		ID:      uuid.New(),
		BookID:  bookID,
		Format:  format,
		FileURL: fileURL,
	}
	result, err := u.bookFileRepo.Create(ctx, bookFile)
	if err != nil {
		return nil, err
	}

	// Auto-count pages for PDF and update the book record.
	if format == entity.BookFileFormatPDF {
		if pages, err := countPDFPages(data); err == nil && pages > 0 {
			if err := u.bookRepo.UpdateTotalPages(ctx, bookID, pages); err != nil {
				slog.WarnContext(ctx, "failed to update total_pages", "book_id", bookID, "err", err)
			}
		} else if err != nil {
			slog.WarnContext(ctx, "failed to count PDF pages", "book_id", bookID, "err", err)
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
	file, err := u.bookFileRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.bookFileRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Reset total_pages on the book when the document file is removed.
	if !file.Format.IsAudio() {
		if err := u.bookRepo.UpdateTotalPages(ctx, file.BookID, 0); err != nil {
			slog.WarnContext(ctx, "failed to reset total_pages after file delete",
				"book_id", file.BookID, "err", err)
		}
	}

	return nil
}

// countPDFPages reads page count from a PDF file in memory.
func countPDFPages(data []byte) (int, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("countPDFPages: %w", err)
	}
	return r.NumPage(), nil
}

// detectContentType returns the correct MIME type for storage metadata.
// http.DetectContentType is unreliable for EPUB (returns application/zip).
func detectContentType(format entity.BookFileFormat, buf []byte) string {
	switch format {
	case entity.BookFileFormatPDF:
		return "application/pdf"
	case entity.BookFileFormatEPUB:
		return "application/epub+zip"
	case entity.BookFileFormatMP3:
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}