package notification

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type notificationUseCase struct {
	repo repository.NotificationRepository
}

func NewNotificationUseCase(repo repository.NotificationRepository) domainUseCase.NotificationUseCase {
	return &notificationUseCase{repo: repo}
}

func (u *notificationUseCase) Create(ctx context.Context, userID uuid.UUID, title, body string) (*entity.Notification, error) {
	n := &entity.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Body:      body,
		IsRead:    false,
		CreatedAt: time.Now(),
	}
	return u.repo.Create(ctx, n)
}

func (u *notificationUseCase) ListMy(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	return u.repo.ListByUser(ctx, userID, limit, offset)
}

func (u *notificationUseCase) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return u.repo.MarkRead(ctx, id, userID)
}

func (u *notificationUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return u.repo.Delete(ctx, id, userID)
}
