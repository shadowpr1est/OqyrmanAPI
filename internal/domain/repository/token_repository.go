package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type TokenRepository interface {
	Save(ctx context.Context, token *entity.Token) error
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entity.Token, error)
	DeleteByRefreshToken(ctx context.Context, refreshToken string) error
	DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error
	// Session management
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Token, error)
	DeleteByID(ctx context.Context, id, userID uuid.UUID) error
}
