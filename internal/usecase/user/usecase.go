package user

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainStorage "github.com/shadowpr1est/OqyrmanAPI/internal/domain/storage"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/phone"
	"golang.org/x/crypto/bcrypt"
)

type userUseCase struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	storage   domainStorage.FileStorage
}

func NewUserUseCase(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, storage domainStorage.FileStorage) domainUseCase.UserUseCase {
	return &userUseCase{userRepo: userRepo, tokenRepo: tokenRepo, storage: storage}
}

func (u *userUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUseCase) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	if user.Email != "" {
		if _, err := mail.ParseAddress(user.Email); err != nil {
			return nil, fmt.Errorf("%w: invalid email format", entity.ErrValidation)
		}
	}
	if user.Phone != "" {
		normalized, err := phone.Normalize(user.Phone)
		if err != nil {
			return nil, err
		}
		user.Phone = normalized
	}
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

func validatePassword(p string) error {
	if len(p) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", entity.ErrValidation)
	}
	var hasUpper, hasDigit bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("%w: password must contain at least one uppercase letter", entity.ErrValidation)
	}
	if !hasDigit {
		return fmt.Errorf("%w: password must contain at least one digit", entity.ErrValidation)
	}
	return nil
}

func (u *userUseCase) CreateStaff(ctx context.Context, email, password, name, surname, rawPhone string, libraryID uuid.UUID) (*entity.UserView, error) {
	if err := validatePassword(password); err != nil {
		return nil, err
	}
	normalizedPhone, err := phone.Normalize(rawPhone)
	if err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("userUseCase.CreateStaff hash: %w", err)
	}

	user := &entity.User{
		ID:            uuid.New(),
		Email:         email,
		PasswordHash:  string(hash),
		Name:          name,
		Surname:       surname,
		Phone:         normalizedPhone,
		Role:          entity.RoleStaff,
		LibraryID:     &libraryID,
		QRCode:        uuid.New().String(),
		EmailVerified: true, // staff создаётся администратором — верификация не нужна
		CreatedAt:     time.Now(),
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

func (u *userUseCase) ListSessions(ctx context.Context, userID uuid.UUID) ([]*entity.Token, error) {
	return u.tokenRepo.ListByUserID(ctx, userID)
}

func (u *userUseCase) RevokeSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	return u.tokenRepo.DeleteByID(ctx, sessionID, userID)
}

func (u *userUseCase) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return u.tokenRepo.DeleteAllByUserID(ctx, userID)
}

func (u *userUseCase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("%w: incorrect current password", entity.ErrValidation)
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return u.userRepo.UpdatePassword(ctx, userID, string(hash))
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
