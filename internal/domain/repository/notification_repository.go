package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) (*entity.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}
