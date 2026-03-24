package reservation

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type reservationUseCase struct {
	reservationRepo repository.ReservationRepository
}

func NewReservationUseCase(
	reservationRepo repository.ReservationRepository,
) domainUseCase.ReservationUseCase {
	return &reservationUseCase{reservationRepo: reservationRepo}
}

func (u *reservationUseCase) Create(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error) {
	if r.LibraryBookID == nil && r.MachineBookID == nil {
		return nil, errors.New("library_book_id or machine_book_id is required")
	}
	r.ID = uuid.New()
	r.Status = entity.ReservationPending
	r.ReservedAt = time.Now()
	return u.reservationRepo.CreateWithDecrement(ctx, r)
}

func (u *reservationUseCase) Return(ctx context.Context, id uuid.UUID) error {
	// Атомарно: pending/active → completed + available_copies + 1
	return u.reservationRepo.ReturnWithIncrement(ctx, id)
}

func (u *reservationUseCase) Cancel(ctx context.Context, id uuid.UUID) error {
	// Было: UpdateStatus(cancelled) — просто меняло статус, available_copies
	// не возвращался. Копия терялась навсегда при отмене брони.
	// Теперь: атомарно pending → cancelled + available_copies + 1
	return u.reservationRepo.CancelWithIncrement(ctx, id)
}

func (u *reservationUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	return u.reservationRepo.GetByID(ctx, id)
}

func (u *reservationUseCase) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListByUser(ctx, userID, limit, offset)
}

func (u *reservationUseCase) ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListAll(ctx, limit, offset, status)
}

func (u *reservationUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	return u.reservationRepo.UpdateStatus(ctx, id, status)
}
