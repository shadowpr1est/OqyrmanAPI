package library

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainStorage "github.com/shadowpr1est/OqyrmanAPI/internal/domain/storage"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
)

type libraryUseCase struct {
	libraryRepo repository.LibraryRepository
	storage     domainStorage.FileStorage
}

func NewLibraryUseCase(
	libraryRepo repository.LibraryRepository,
	storage domainStorage.FileStorage,
) domainUseCase.LibraryUseCase {
	return &libraryUseCase{libraryRepo: libraryRepo, storage: storage}
}

func (u *libraryUseCase) Create(ctx context.Context, library *entity.Library) (*entity.Library, error) {
	library.ID = uuid.New()
	return u.libraryRepo.Create(ctx, library)
}

func (u *libraryUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Library, error) {
	return u.libraryRepo.GetByID(ctx, id)
}

func (u *libraryUseCase) List(ctx context.Context, limit, offset int) ([]*entity.Library, int, error) {
	return u.libraryRepo.List(ctx, limit, offset)
}

func (u *libraryUseCase) ListNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.Library, error) {
	return u.libraryRepo.ListNearby(ctx, lat, lng, radiusKm)
}

func (u *libraryUseCase) Update(ctx context.Context, library *entity.Library) (*entity.Library, error) {
	return u.libraryRepo.Update(ctx, library)
}

func (u *libraryUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.libraryRepo.Delete(ctx, id)
}

func (u *libraryUseCase) UploadPhoto(ctx context.Context, id uuid.UUID, photo *fileupload.File) (*entity.Library, error) {
	objectKey := fmt.Sprintf("libraries/%s/photo", id.String())
	url, err := u.storage.Upload(ctx, objectKey, photo.Reader, photo.Size, photo.ContentType)
	if err != nil {
		return nil, fmt.Errorf("libraryUseCase.UploadPhoto upload: %w", err)
	}
	if err := u.libraryRepo.UpdatePhotoURL(ctx, id, url); err != nil {
		return nil, err
	}
	return u.libraryRepo.GetByID(ctx, id)
}
