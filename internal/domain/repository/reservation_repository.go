package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReservationRepository interface {
	// CreateWithDecrement атомарно проверяет наличие копий, уменьшает
	// available_copies и создаёт бронь в одной транзакции.
	CreateWithDecrement(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error)

	// ReturnWithIncrement атомарно переводит бронь pending/active → completed,
	// проставляет returned_at и увеличивает available_copies на 1.
	// Возвращает ошибку если текущий статус не допускает возврат.
	ReturnWithIncrement(ctx context.Context, id uuid.UUID) error

	// CancelWithIncrement атомарно переводит бронь pending → cancelled
	// и увеличивает available_copies на 1.
	// Возвращает ошибку если текущий статус не pending.
	CancelWithIncrement(ctx context.Context, id uuid.UUID) error

	GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Reservation, error)
	ListAll(ctx context.Context, limit, offset int) ([]*entity.Reservation, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}
