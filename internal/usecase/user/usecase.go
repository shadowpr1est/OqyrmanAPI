package user

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainStorage "github.com/shadowpr1est/OqyrmanAPI/internal/domain/storage"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
	"golang.org/x/crypto/bcrypt"
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

func (u *userUseCase) AdminUpdateUser(
	ctx context.Context,
	id uuid.UUID,
	role *entity.Role,
	libraryID *uuid.UUID,
	name, surname, email, phone *string,
) (*entity.UserView, error) {
	// Валидация: Staff требует library_id, остальные — нет
	if role != nil {
		if *role == entity.RoleStaff && libraryID == nil {
			return nil, errors.New("library_id is required for Staff role")
		}
		if *role != entity.RoleStaff && libraryID != nil {
			return nil, errors.New("library_id must be null for non-Staff roles")
		}
	}

	return u.userRepo.AdminUpdate(ctx, id, role, libraryID, name, surname, email, phone)
}

func (u *userUseCase) AdminDelete(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}

func (u *userUseCase) ListAllView(ctx context.Context, limit, offset int) ([]*entity.UserView, int, error) {
	return u.userRepo.ListAllView(ctx, limit, offset)
}

func (u *userUseCase) CreateStaff(ctx context.Context, email, password, name, surname, phone string, libraryID uuid.UUID) (*entity.UserView, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("userUseCase.CreateStaff hash: %w", err)
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Surname:      surname,
		FullName:     name + " " + surname,
		Phone:        phone,
		Role:         entity.RoleStaff,
		LibraryID:    &libraryID,
		QRCode:       uuid.New().String(),
		CreatedAt:    time.Now(),
	}

	if _, err := u.userRepo.Create(ctx, user); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, entity.ErrEmailTaken
		}
		return nil, fmt.Errorf("userUseCase.CreateStaff: %w", err)
	}

	// Возвращаем UserView с именем библиотеки
	return u.userRepo.AdminUpdate(ctx, user.ID, nil, nil, nil, nil, nil, nil)
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
