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
	ReturnWithIncrement(ctx context.Context, id uuid.UUID) error

	// CancelWithIncrement атомарно переводит бронь pending → cancelled
	// и увеличивает available_copies на 1.
	CancelWithIncrement(ctx context.Context, id uuid.UUID) error

	// CancelOverdue находит все брони где due_date < now() AND status = 'active',
	// переводит в cancelled и восстанавливает available_copies.
	// Возвращает количество отменённых броней.
	CancelOverdue(ctx context.Context) (int, error)

	GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error)

	// ListByUser возвращает брони пользователя с пагинацией.
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error)

	// ListAll возвращает все брони с пагинацией.
	// status — необязательный фильтр: nil = все брони, значение = фильтр по статусу.
	ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error)

	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}
