package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthUseCase interface {
	Register(ctx context.Context, user *entity.User) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
}
