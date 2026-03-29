package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	userRepo       repository.UserRepository
	tokenRepo      repository.TokenRepository
	jwt            *jwt.Manager
	refreshTokenTTL time.Duration
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwt *jwt.Manager,
	refreshTokenTTLDays int,
) domainUseCase.AuthUseCase {
	return &authUseCase{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		jwt:             jwt,
		refreshTokenTTL: time.Duration(refreshTokenTTLDays) * 24 * time.Hour,
	}
}

func (u *authUseCase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	existing, err := u.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, entity.ErrNotFound) {
		return nil, fmt.Errorf("authUseCase.Register lookup: %w", err)
	}
	if existing != nil {
		return nil, entity.ErrEmailTaken
	}

	if err := validateEmail(user.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(user.PasswordHash); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user.ID = uuid.New()
	user.PasswordHash = string(hash)
	user.Role = entity.RoleUser
	user.CreatedAt = time.Now()
	// FIX: QR-код не генерировался — поле всегда было пустым.
	// QR содержит userID — достаточно для идентификации читателя.
	// Фронт рендерит QR-картинку из этой строки через любую QR-библиотеку.
	user.QRCode = user.ID.String()

	return u.userRepo.Create(ctx, user)
}

func (u *authUseCase) Login(ctx context.Context, email, password string) (*domainUseCase.TokenPair, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := u.jwt.GenerateAccessToken(user.ID, string(user.Role), user.LibraryID)
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	token := &entity.Token{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(u.refreshTokenTTL),
		CreatedAt:    time.Now(),
	}

	if err := u.tokenRepo.Save(ctx, token); err != nil {
		return nil, err
	}

	return &domainUseCase.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	return u.tokenRepo.DeleteByRefreshToken(ctx, refreshToken)
}

func validateEmail(email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("%w: invalid email format", entity.ErrValidation)
	}
	return nil
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

func (u *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domainUseCase.TokenPair, error) {
	token, err := u.tokenRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	user, err := u.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	if err := u.tokenRepo.DeleteByRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	accessToken, err := u.jwt.GenerateAccessToken(user.ID, string(user.Role), user.LibraryID)
	if err != nil {
		return nil, err
	}

	newRefreshToken := uuid.New().String()
	newToken := &entity.Token{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(u.refreshTokenTTL),
		CreatedAt:    time.Now(),
	}

	if err := u.tokenRepo.Save(ctx, newToken); err != nil {
		return nil, err
	}

	return &domainUseCase.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
