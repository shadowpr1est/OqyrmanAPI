package author

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainStorage "github.com/shadowpr1est/OqyrmanAPI/internal/domain/storage"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type authorUseCase struct {
	authorRepo repository.AuthorRepository
	storage    domainStorage.FileStorage
}

func NewAuthorUseCase(authorRepo repository.AuthorRepository, storage domainStorage.FileStorage) domainUseCase.AuthorUseCase {
	return &authorUseCase{authorRepo: authorRepo, storage: storage}
}

func (u *authorUseCase) Create(ctx context.Context, author *entity.Author) (*entity.Author, error) {
	author.ID = uuid.New()
	return u.authorRepo.Create(ctx, author)
}

func (u *authorUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Author, error) {
	return u.authorRepo.GetByID(ctx, id)
}

func (u *authorUseCase) List(ctx context.Context, limit, offset int) ([]*entity.Author, int, error) {
	return u.authorRepo.List(ctx, limit, offset)
}

func (u *authorUseCase) Update(ctx context.Context, author *entity.Author) (*entity.Author, error) {
	return u.authorRepo.Update(ctx, author)
}

func (u *authorUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.authorRepo.Delete(ctx, id)
}

func (u *authorUseCase) Search(ctx context.Context, query string, limit, offset int) ([]*entity.Author, int, error) {
	return u.authorRepo.Search(ctx, query, limit, offset)
}

func (u *authorUseCase) UploadPhoto(ctx context.Context, id uuid.UUID, file io.Reader, size int64, contentType string) (*entity.Author, error) {
	objectKey := fmt.Sprintf("authors/%s/photo", id.String())
	url, err := u.storage.Upload(ctx, objectKey, file, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("authorUseCase.UploadPhoto upload: %w", err)
	}
	if err := u.authorRepo.UpdatePhotoURL(ctx, id, url); err != nil {
		return nil, err
	}
	return u.authorRepo.GetByID(ctx, id)
}
