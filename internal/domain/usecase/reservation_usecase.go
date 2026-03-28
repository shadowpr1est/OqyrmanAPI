package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReservationUseCase interface {
	Create(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error)

	// User
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error)
	Cancel(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error

	// Staff — только своя библиотека
	ListByLibrary(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.Reservation, int, error)
	StaffCancel(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error
	StaffReturn(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error
	StaffUpdateStatus(ctx context.Context, id uuid.UUID, libraryID uuid.UUID, status entity.ReservationStatus) error

	Extend(ctx context.Context, id, userID uuid.UUID, newDueDate time.Time) (*entity.Reservation, error)

	// Admin — без ограничений
	ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error)
	AdminReturn(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error
}
