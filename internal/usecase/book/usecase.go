package book

import (
	"context"
	"errors"
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

type bookUseCase struct {
	bookRepo repository.BookRepository
	storage  domainStorage.FileStorage
}

func NewBookUseCase(bookRepo repository.BookRepository, storage domainStorage.FileStorage) domainUseCase.BookUseCase {
	return &bookUseCase{bookRepo: bookRepo, storage: storage}
}

func (u *bookUseCase) Create(ctx context.Context, book *entity.Book, cover *fileupload.File) (*entity.Book, error) {
	book.ID = uuid.New()

	if cover != nil {
		if u.storage == nil {
			return nil, errors.New("file storage is not configured")
		}
		coverURL, err := u.uploadCover(ctx, book.ID, cover)
		if err != nil {
			return nil, err
		}
		book.CoverURL = coverURL
	}

	return u.bookRepo.Create(ctx, book)
}

func (u *bookUseCase) UploadCover(ctx context.Context, id uuid.UUID, cover *fileupload.File) (*entity.Book, error) {
	if u.storage == nil {
		return nil, errors.New("file storage is not configured")
	}
	coverURL, err := u.uploadCover(ctx, id, cover)
	if err != nil {
		return nil, err
	}
	if err := u.bookRepo.UpdateCoverURL(ctx, id, coverURL); err != nil {
		return nil, err
	}
	return u.bookRepo.GetByID(ctx, id)
}

// приватный хелпер — логика загрузки не дублируется
func (u *bookUseCase) uploadCover(ctx context.Context, id uuid.UUID, cover *fileupload.File) (string, error) {
	ext := strings.ToLower(filepath.Ext(cover.Filename))
	objectKey := fmt.Sprintf("covers/%s%s", id.String(), ext)
	return u.storage.Upload(ctx, objectKey, cover.Reader, cover.Size, cover.ContentType)
}

func (u *bookUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Book, error) {
	return u.bookRepo.GetByID(ctx, id)
}

func (u *bookUseCase) List(ctx context.Context, limit, offset int) ([]*entity.Book, int, error) {
	return u.bookRepo.List(ctx, limit, offset)
}

func (u *bookUseCase) ListByAuthor(ctx context.Context, authorID uuid.UUID) ([]*entity.Book, error) {
	return u.bookRepo.ListByAuthor(ctx, authorID)
}

func (u *bookUseCase) ListByGenre(ctx context.Context, genreID uuid.UUID) ([]*entity.Book, error) {
	return u.bookRepo.ListByGenre(ctx, genreID)
}

func (u *bookUseCase) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Book, int, error) {
	return u.bookRepo.Search(ctx, query, limit, offset)
}

func (u *bookUseCase) Update(ctx context.Context, book *entity.Book) (*entity.Book, error) {
	return u.bookRepo.Update(ctx, book)
}

func (u *bookUseCase) ListPopular(ctx context.Context, limit, offset int) ([]*entity.Book, int, error) {
	return u.bookRepo.ListPopular(ctx, limit, offset)
}

func (u *bookUseCase) ListSimilar(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.Book, error) {
	return u.bookRepo.ListSimilar(ctx, bookID, limit)
}

func (u *bookUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.bookRepo.Delete(ctx, id)
}

func (u *bookUseCase) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.BookView, error) {
	return u.bookRepo.GetByIDView(ctx, id)
}

func (u *bookUseCase) ListView(ctx context.Context, limit, offset int) ([]*entity.BookView, int, error) {
	return u.bookRepo.ListView(ctx, limit, offset)
}

func (u *bookUseCase) ListByAuthorView(ctx context.Context, authorID uuid.UUID) ([]*entity.BookView, error) {
	return u.bookRepo.ListByAuthorView(ctx, authorID)
}

func (u *bookUseCase) ListByGenreView(ctx context.Context, genreID uuid.UUID) ([]*entity.BookView, error) {
	return u.bookRepo.ListByGenreView(ctx, genreID)
}

func (u *bookUseCase) SearchView(ctx context.Context, query string, limit, offset int) ([]*entity.BookView, int, error) {
	return u.bookRepo.SearchView(ctx, query, limit, offset)
}

func (u *bookUseCase) ListPopularView(ctx context.Context, limit, offset int) ([]*entity.BookView, int, error) {
	return u.bookRepo.ListPopularView(ctx, limit, offset)
}

func (u *bookUseCase) ListSimilarView(ctx context.Context, bookID uuid.UUID, limit int) ([]*entity.BookView, error) {
	return u.bookRepo.ListSimilarView(ctx, bookID, limit)
}
