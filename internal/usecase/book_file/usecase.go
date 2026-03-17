package book_file

import (
	"context"
	"fmt"
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
	storage      domainStorage.FileStorage
}

func NewBookFileUseCase(bookFileRepo repository.BookFileRepository, storage domainStorage.FileStorage) domainUseCase.BookFileUseCase {
	return &bookFileUseCase{bookFileRepo: bookFileRepo, storage: storage}
}

func (u *bookFileUseCase) Create(ctx context.Context, file *entity.BookFile) (*entity.BookFile, error) {
	file.ID = uuid.New()
	return u.bookFileRepo.Create(ctx, file)
}

func (u *bookFileUseCase) Upload(ctx context.Context, bookID uuid.UUID, format string, isAudio bool, file *fileupload.File) (*entity.BookFile, error) {
	if u.storage == nil {
		return nil, fmt.Errorf("file storage is not configured")
	}
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
