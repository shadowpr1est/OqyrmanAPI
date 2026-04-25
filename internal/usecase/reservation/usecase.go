package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

// Broadcaster is satisfied by *hub.NotificationHub.
type Broadcaster interface {
	Send(userID uuid.UUID, n *entity.Notification)
}

type reservationUseCase struct {
	reservationRepo repository.ReservationRepository
	userRepo        repository.UserRepository
	notifRepo       repository.NotificationRepository
	hub             Broadcaster // optional — nil disables SSE push
}

func NewReservationUseCase(
	repo repository.ReservationRepository,
	userRepo repository.UserRepository,
	notifRepo repository.NotificationRepository,
	hub Broadcaster,
) domainUseCase.ReservationUseCase {
	return &reservationUseCase{
		reservationRepo: repo,
		userRepo:        userRepo,
		notifRepo:       notifRepo,
		hub:             hub,
	}
}

func (u *reservationUseCase) notify(ctx context.Context, userID uuid.UUID, nType entity.NotificationType, title, body string) {
	n := &entity.Notification{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      nType,
		Title:     title,
		Body:      body,
		CreatedAt: time.Now(),
	}
	saved, err := u.notifRepo.Create(ctx, n)
	if err != nil {
		slog.ErrorContext(ctx, "failed to create notification", "user_id", userID, "err", err)
		return
	}
	if u.hub != nil {
		u.hub.Send(userID, saved)
	}
}

func (u *reservationUseCase) Create(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error) {
	r.ID = uuid.New()
	r.Status = entity.ReservationPending
	r.ReservedAt = time.Now()
	res, err := u.reservationRepo.CreateWithDecrement(ctx, r)
	if err != nil {
		return nil, err
	}
	u.notify(ctx, res.UserID, entity.NotifReservationSuccess, "Бронирование создано",
		"Ваше бронирование принято и ожидает подтверждения.")
	return res, nil
}

func (u *reservationUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	return u.reservationRepo.GetByID(ctx, id)
}

// --- User ---

func (u *reservationUseCase) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListByUser(ctx, userID, limit, offset)
}

func (u *reservationUseCase) Cancel(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	if err := u.reservationRepo.CancelWithIncrement(ctx, id, callerID); err != nil {
		return err
	}
	u.notify(ctx, callerID, entity.NotifReservationExpired, "Бронирование отменено",
		"Ваше бронирование было отменено.")
	return nil
}

func (u *reservationUseCase) Extend(ctx context.Context, id, userID uuid.UUID) (*entity.Reservation, error) {
	res, err := u.reservationRepo.Extend(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	u.notify(ctx, userID, entity.NotifReservationSuccess, "Срок бронирования продлён",
		"Срок возврата книги продлён до "+res.DueDate.Format("02.01.2006")+".")
	return res, nil
}

// --- Staff ---

func (u *reservationUseCase) ListByLibrary(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListByLibrary(ctx, libraryID, limit, offset, status)
}

func (u *reservationUseCase) StaffCancel(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	r, err := u.reservationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.reservationRepo.StaffCancel(ctx, id, libraryID); err != nil {
		return err
	}
	u.notify(ctx, r.UserID, entity.NotifReservationExpired, "Бронирование отменено библиотекой",
		"Ваше бронирование было отменено сотрудником библиотеки.")
	return nil
}

func (u *reservationUseCase) StaffReturn(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	r, err := u.reservationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.reservationRepo.StaffReturn(ctx, id, libraryID); err != nil {
		return err
	}
	u.notify(ctx, r.UserID, entity.NotifGeneral, "Книга возвращена",
		"Возврат книги зафиксирован. Спасибо!")
	return nil
}

// --- QR Scan ---

func (u *reservationUseCase) LookupUserByQR(ctx context.Context, qrCode string, libraryID uuid.UUID) (*entity.User, []*entity.ReservationView, error) {
	user, err := u.userRepo.GetByQRCode(ctx, qrCode)
	if err != nil {
		return nil, nil, err
	}
	reservations, err := u.reservationRepo.ListPendingByUserAndLibraryView(ctx, user.ID, libraryID)
	if err != nil {
		return nil, nil, err
	}
	return user, reservations, nil
}

func (u *reservationUseCase) ScanQR(ctx context.Context, qrToken string, libraryID uuid.UUID) (*entity.ReservationView, error) {
	res, err := u.reservationRepo.ActivateByQRToken(ctx, qrToken, libraryID)
	if err != nil {
		return nil, err
	}
	u.notify(ctx, res.UserID, entity.NotifPickupSuccess, "Книга выдана",
		"Книга выдана вам на 30 дней. Срок возврата: "+res.DueDate.Format("02.01.2006")+".")
	return u.reservationRepo.GetByIDView(ctx, res.ID)
}

// --- Admin ---

func (u *reservationUseCase) ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	return u.reservationRepo.ListAll(ctx, limit, offset, status)
}

func (u *reservationUseCase) AdminReturn(ctx context.Context, id uuid.UUID) error {
	r, err := u.reservationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := u.reservationRepo.AdminReturn(ctx, id); err != nil {
		return err
	}
	u.notify(ctx, r.UserID, entity.NotifGeneral, "Книга возвращена",
		"Возврат книги зафиксирован администратором.")
	return nil
}

func (u *reservationUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	if err := u.reservationRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}
	if status == entity.ReservationActive {
		if r, err := u.reservationRepo.GetByID(ctx, id); err == nil {
			u.notify(ctx, r.UserID, entity.NotifPickupSuccess, "Книга выдана",
				"Книга выдана вам на 30 дней. Срок возврата: "+r.DueDate.Format("02.01.2006")+".")
		}
	}
	return nil
}

func (u *reservationUseCase) StaffUpdateStatus(ctx context.Context, id uuid.UUID, libraryID uuid.UUID, status entity.ReservationStatus) error {
	if err := u.reservationRepo.StaffUpdateStatus(ctx, id, libraryID, status); err != nil {
		return err
	}
	if status == entity.ReservationActive {
		if r, err := u.reservationRepo.GetByID(ctx, id); err == nil {
			u.notify(ctx, r.UserID, entity.NotifPickupSuccess, "Книга выдана",
				"Книга выдана вам на 30 дней. Срок возврата: "+r.DueDate.Format("02.01.2006")+".")
		}
	}
	return nil
}

func (u *reservationUseCase) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReservationView, error) {
	return u.reservationRepo.GetByIDView(ctx, id)
}

func (u *reservationUseCase) ListByUserView(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.ReservationView, int, error) {
	return u.reservationRepo.ListByUserView(ctx, userID, limit, offset)
}

func (u *reservationUseCase) ListByLibraryView(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.ReservationView, int, error) {
	return u.reservationRepo.ListByLibraryView(ctx, libraryID, limit, offset, status)
}

func (u *reservationUseCase) ListAllView(ctx context.Context, limit, offset int, status *string) ([]*entity.ReservationView, int, error) {
	return u.reservationRepo.ListAllView(ctx, limit, offset, status)
}
