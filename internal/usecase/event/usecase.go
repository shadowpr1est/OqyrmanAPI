package event

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainStorage "github.com/shadowpr1est/OqyrmanAPI/internal/domain/storage"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
)

type eventUseCase struct {
	repo    repository.EventRepository
	storage domainStorage.FileStorage
}

func NewEventUseCase(repo repository.EventRepository, storage domainStorage.FileStorage) domainUseCase.EventUseCase {
	return &eventUseCase{repo: repo, storage: storage}
}

func (u *eventUseCase) List(ctx context.Context, limit, offset int) ([]*entity.Event, int, error) {
	return u.repo.List(ctx, limit, offset)
}

func (u *eventUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *eventUseCase) Create(ctx context.Context, e *entity.Event, cover *fileupload.File) (*entity.Event, error) {
	e.ID = uuid.New()
	e.CreatedAt = time.Now()
	if cover != nil {
		url, err := u.uploadCover(ctx, e.ID, cover)
		if err != nil {
			return nil, err
		}
		e.CoverURL = &url
	}
	return u.repo.Create(ctx, e)
}

func (u *eventUseCase) Update(ctx context.Context, e *entity.Event, cover *fileupload.File) (*entity.Event, error) {
	if cover != nil {
		url, err := u.uploadCover(ctx, e.ID, cover)
		if err != nil {
			return nil, err
		}
		e.CoverURL = &url
	}
	return u.repo.Update(ctx, e)
}

func (u *eventUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}

func (u *eventUseCase) uploadCover(ctx context.Context, id uuid.UUID, cover *fileupload.File) (string, error) {
	if u.storage == nil {
		return "", errors.New("file storage is not configured")
	}
	ext := strings.ToLower(filepath.Ext(cover.Filename))
	objectKey := fmt.Sprintf("events/%s%s", id.String(), ext)
	return u.storage.Upload(ctx, objectKey, cover.Reader, cover.Size, cover.ContentType)
}
