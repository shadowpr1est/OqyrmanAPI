package usecase

import (
	"context"

	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	User         *entity.User
}

type AuthUseCase interface {
	Register(ctx context.Context, user *entity.User) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	// SendVerificationCode генерирует и отправляет 6-значный код на email пользователя.
	SendVerificationCode(ctx context.Context, email string) error
	// VerifyEmail проверяет код и возвращает токены.
	VerifyEmail(ctx context.Context, email, code string) (*TokenPair, error)
	// ForgotPassword отправляет код сброса пароля на email.
	ForgotPassword(ctx context.Context, email string) error
	// ResendResetCode повторно отправляет код сброса пароля.
	ResendResetCode(ctx context.Context, email string) error
	// ResetPassword проверяет код и устанавливает новый пароль.
	ResetPassword(ctx context.Context, email, code, newPassword string) error
	// LoginWithGoogle верифицирует Google ID token и возвращает JWT токены.
	LoginWithGoogle(ctx context.Context, idToken string) (*TokenPair, error)
}
