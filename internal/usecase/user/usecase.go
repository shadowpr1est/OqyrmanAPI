package user

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

type userUseCase struct {
	userRepo repository.UserRepository
	storage  domainStorage.FileStorage
}

func NewUserUseCase(userRepo repository.UserRepository, storage domainStorage.FileStorage) domainUseCase.UserUseCase {
	return &userUseCase{userRepo: userRepo, storage: storage}
}

func (u *userUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUseCase) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	return u.userRepo.Update(ctx, user)
}

func (u *userUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}

func (u *userUseCase) ListAll(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	return u.userRepo.ListAll(ctx, limit, offset)
}

func (u *userUseCase) UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role, libraryID *uuid.UUID) error {
	// libraryID обязателен для Staff, должен быть nil для Admin и User
	if role == entity.RoleStaff && libraryID == nil {
		return errors.New("library_id is required for Staff role")
	}
	if role != entity.RoleStaff && libraryID != nil {
		return errors.New("library_id must be null for non-Staff roles")
	}
	return u.userRepo.UpdateRole(ctx, id, role, libraryID)
}

func (u *userUseCase) AdminDelete(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}

func (u *userUseCase) ListAllView(ctx context.Context, limit, offset int) ([]*entity.UserView, int, error) {
	return u.userRepo.ListAllView(ctx, limit, offset)
}

func (u *userUseCase) UploadAvatar(ctx context.Context, id uuid.UUID, avatar *fileupload.File) (*entity.User, error) {
	if u.storage == nil {
		return nil, errors.New("file storage is not configured")
	}
	ext := strings.ToLower(filepath.Ext(avatar.Filename))
	objectKey := fmt.Sprintf("avatars/%s%s", id.String(), ext)

	avatarURL, err := u.storage.Upload(ctx, objectKey, avatar.Reader, avatar.Size, avatar.ContentType)
	if err != nil {
		return nil, err
	}
	if err := u.userRepo.UpdateAvatarURL(ctx, id, avatarURL); err != nil {
		return nil, err
	}
	return u.userRepo.GetByID(ctx, id)
}
