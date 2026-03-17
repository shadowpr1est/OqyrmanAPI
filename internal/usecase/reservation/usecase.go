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
	libBookRepo     repository.LibraryBookRepository     // NEW
	machBookRepo    repository.BookMachineBookRepository // NEW
}

func NewReservationUseCase(
	reservationRepo repository.ReservationRepository,
	libBookRepo repository.LibraryBookRepository,
	machBookRepo repository.BookMachineBookRepository,
) domainUseCase.ReservationUseCase {
	return &reservationUseCase{reservationRepo, libBookRepo, machBookRepo}
}

func (u *reservationUseCase) Return(ctx context.Context, id uuid.UUID) error {
	r, err := u.reservationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now()
	r.ReturnedAt = &now
	r.Status = entity.ReservationCompleted
	if err := u.reservationRepo.UpdateStatus(ctx, id, entity.ReservationCompleted); err != nil {
		return err
	}
	// восстановить available_copies
	if r.LibraryBookID != nil {
		lb, err := u.libBookRepo.GetByID(ctx, *r.LibraryBookID)
		if err != nil {
			return err
		}
		lb.AvailableCopies++
		_, err = u.libBookRepo.Update(ctx, lb)
		return err
	}
	if r.MachineBookID != nil {
		mb, err := u.machBookRepo.GetByID(ctx, *r.MachineBookID)
		if err != nil {
			return err
		}
		mb.AvailableCopies++
		_, err = u.machBookRepo.Update(ctx, mb)
		return err
	}
	return nil
}

func (u *reservationUseCase) Create(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error) {
	if r.LibraryBookID == nil && r.MachineBookID == nil {
		return nil, errors.New("library_book_id or machine_book_id is required")
	}
	r.ID = uuid.New()
	r.Status = entity.ReservationPending
	r.ReservedAt = time.Now()
	return u.reservationRepo.Create(ctx, r)
}

func (u *reservationUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	return u.reservationRepo.GetByID(ctx, id)
}

func (u *reservationUseCase) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Reservation, error) {
	return u.reservationRepo.ListByUser(ctx, userID)
}

func (u *reservationUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	return u.reservationRepo.UpdateStatus(ctx, id, status)
}

func (u *reservationUseCase) Cancel(ctx context.Context, id uuid.UUID) error {
	return u.reservationRepo.UpdateStatus(ctx, id, entity.ReservationCancelled)
}
func (u *reservationUseCase) ListAll(ctx context.Context, limit, offset int) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListAll(ctx, limit, offset)
}
