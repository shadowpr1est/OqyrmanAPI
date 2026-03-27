package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReservationRepository interface {
	CreateWithDecrement(ctx context.Context, res *entity.Reservation) (*entity.Reservation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error)

	// Пользователь видит только свои
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error)

	// Admin — всё без фильтрации
	ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error)
	AdminReturn(ctx context.Context, id uuid.UUID) error

	// Staff — только своя библиотека
	ListByLibrary(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.Reservation, int, error)
	StaffCancel(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error
	StaffReturn(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error
	// callerID != nil → проверка владельца (user cancel)
	// libraryID != nil → проверка принадлежности библиотеке (staff cancel/return)
	// оба nil → без проверки (admin)
	CancelWithIncrement(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error

	Extend(ctx context.Context, id, userID uuid.UUID, newDueDate time.Time) (*entity.Reservation, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error
	CancelOverdue(ctx context.Context) (int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
