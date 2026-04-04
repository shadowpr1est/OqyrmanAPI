package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type PasswordResetCodeRepository interface {
	Save(ctx context.Context, code *entity.PasswordResetCode) error
	// GetByUserID возвращает актуальный код для пользователя (для сравнения хеша в usecase).
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.PasswordResetCode, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
