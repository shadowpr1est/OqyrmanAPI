package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	userRepo  repository.UserRepository
	tokenRepo repository.TokenRepository
	jwt       *jwt.Manager
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwt *jwt.Manager,
) domainUseCase.AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwt:       jwt,
	}
}

func (u *authUseCase) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	existing, _ := u.userRepo.GetByEmail(ctx, user.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
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

	accessToken, err := u.jwt.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	token := &entity.Token{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 30),
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

func (u *authUseCase) Logout(ctx context.Context, userID uuid.UUID) error {
	return u.tokenRepo.DeleteAllByUserID(ctx, userID)
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

	accessToken, err := u.jwt.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	newRefreshToken := uuid.New().String()
	newToken := &entity.Token{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 30),
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

func (u *authUseCase) Me(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, userID)
}
