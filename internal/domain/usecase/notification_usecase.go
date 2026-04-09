package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type NotificationUseCase interface {
	Create(ctx context.Context, userID uuid.UUID, nType entity.NotificationType, title, body string) (*entity.Notification, error)
	ListMy(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
