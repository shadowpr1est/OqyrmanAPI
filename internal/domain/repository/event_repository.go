package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type EventRepository interface {
	List(ctx context.Context, limit, offset int) ([]*entity.Event, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Event, error)
	Create(ctx context.Context, e *entity.Event) (*entity.Event, error)
	Update(ctx context.Context, e *entity.Event) (*entity.Event, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindUpcoming(ctx context.Context, lookahead time.Duration) ([]*entity.Event, error)
	MarkReminderSent(ctx context.Context, id uuid.UUID) error
}
