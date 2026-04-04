package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type PasswordResetCodeRepository interface {
	Save(ctx context.Context, code *entity.PasswordResetCode) error
	GetByUserAndCode(ctx context.Context, userID uuid.UUID, code string) (*entity.PasswordResetCode, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
