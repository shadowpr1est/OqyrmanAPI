package reservation

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type reservationUseCase struct {
	reservationRepo repository.ReservationRepository
}

func NewReservationUseCase(repo repository.ReservationRepository) domainUseCase.ReservationUseCase {
	return &reservationUseCase{reservationRepo: repo}
}

func (u *reservationUseCase) Create(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error) {
	r.ID = uuid.New()
	r.Status = entity.ReservationPending
	r.ReservedAt = time.Now()
	return u.reservationRepo.CreateWithDecrement(ctx, r)
}

func (u *reservationUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	return u.reservationRepo.GetByID(ctx, id)
}

// --- User ---

func (u *reservationUseCase) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListByUser(ctx, userID, limit, offset)
}

func (u *reservationUseCase) Cancel(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	return u.reservationRepo.CancelWithIncrement(ctx, id, callerID)
}

// --- Staff ---

func (u *reservationUseCase) ListByLibrary(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListByLibrary(ctx, libraryID, limit, offset, status)
}

func (u *reservationUseCase) StaffCancel(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	return u.reservationRepo.StaffCancel(ctx, id, libraryID)
}

func (u *reservationUseCase) StaffReturn(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	return u.reservationRepo.StaffReturn(ctx, id, libraryID)
}

// --- Admin ---

func (u *reservationUseCase) ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListAll(ctx, limit, offset, status)
}

func (u *reservationUseCase) AdminReturn(ctx context.Context, id uuid.UUID) error {
	return u.reservationRepo.AdminReturn(ctx, id)
}

func (u *reservationUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	return u.reservationRepo.UpdateStatus(ctx, id, status)
}
