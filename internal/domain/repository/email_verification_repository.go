package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type EmailVerificationCodeRepository interface {
	// Save создаёт или перезаписывает код для пользователя (UPSERT по user_id).
	Save(ctx context.Context, code *entity.EmailVerificationCode) error
	// GetByUserAndCode возвращает код, если он принадлежит пользователю.
	GetByUserAndCode(ctx context.Context, userID uuid.UUID, code string) (*entity.EmailVerificationCode, error)
	// DeleteByUserID удаляет все коды для пользователя (после успешной верификации).
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
