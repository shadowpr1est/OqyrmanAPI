package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReservationUseCase interface {
	Create(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Reservation, error)
	// ListAll — status необязателен: nil = все брони, значение = фильтр по статусу.
	ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error
	Cancel(ctx context.Context, id uuid.UUID) error
	Return(ctx context.Context, id uuid.UUID) error
}
