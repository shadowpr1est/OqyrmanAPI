package usecase

import (
	"context"

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

	Extend(ctx context.Context, id, userID uuid.UUID) (*entity.Reservation, error)

	// QR scan — staff activates reservation by scanning user's QR
	ScanQR(ctx context.Context, qrToken string, libraryID uuid.UUID) (*entity.ReservationView, error)

	// Admin — без ограничений
	ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error)
	AdminReturn(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error

	// View methods — return enriched nested data for GET endpoints.
	GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReservationView, error)
	ListByUserView(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.ReservationView, int, error)
	ListByLibraryView(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.ReservationView, int, error)
	ListAllView(ctx context.Context, limit, offset int, status *string) ([]*entity.ReservationView, int, error)
}
