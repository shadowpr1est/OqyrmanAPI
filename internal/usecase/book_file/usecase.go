package book_file

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainStorage "github.com/shadowpr1est/OqyrmanAPI/internal/domain/storage"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type bookFileUseCase struct {
	bookFileRepo repository.BookFileRepository
	storage      domainStorage.FileStorage
}

func NewBookFileUseCase(bookFileRepo repository.BookFileRepository, storage domainStorage.FileStorage) domainUseCase.BookFileUseCase {
	return &bookFileUseCase{bookFileRepo: bookFileRepo, storage: storage}
}

func (u *bookFileUseCase) Create(ctx context.Context, file *entity.BookFile) (*entity.BookFile, error) {
	file.ID = uuid.New()
	return u.bookFileRepo.Create(ctx, file)
}

func (u *bookFileUseCase) Upload(ctx context.Context, bookID uuid.UUID, format string, isAudio bool, filename string, reader io.Reader, size int64, contentType string) (*entity.BookFile, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	objectKey := fmt.Sprintf("books/%s/%s%s", bookID.String(), uuid.New().String(), ext)

	fileURL, err := u.storage.Upload(ctx, objectKey, reader, size, contentType)
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
	return u.bookFileRepo.Create(ctx, bookFile)
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
