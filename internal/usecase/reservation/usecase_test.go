package reservation_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/internal/usecase/reservation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mocks ────────────────────────────────────────────────────────────────────

type mockReservationRepo struct{ mock.Mock }

func (m *mockReservationRepo) CreateWithDecrement(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error) {
	args := m.Called(ctx, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Reservation), args.Error(1)
}
func (m *mockReservationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Reservation), args.Error(1)
}
func (m *mockReservationRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entity.Reservation), args.Int(1), args.Error(2)
}
func (m *mockReservationRepo) ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	args := m.Called(ctx, limit, offset, status)
	return args.Get(0).([]*entity.Reservation), args.Int(1), args.Error(2)
}
func (m *mockReservationRepo) AdminReturn(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockReservationRepo) ListByLibrary(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	args := m.Called(ctx, libraryID, limit, offset, status)
	return args.Get(0).([]*entity.Reservation), args.Int(1), args.Error(2)
}
func (m *mockReservationRepo) StaffCancel(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	return m.Called(ctx, id, libraryID).Error(0)
}
func (m *mockReservationRepo) StaffReturn(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	return m.Called(ctx, id, libraryID).Error(0)
}
func (m *mockReservationRepo) StaffUpdateStatus(ctx context.Context, id uuid.UUID, libraryID uuid.UUID, status entity.ReservationStatus) error {
	return m.Called(ctx, id, libraryID, status).Error(0)
}
func (m *mockReservationRepo) CancelWithIncrement(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	return m.Called(ctx, id, callerID).Error(0)
}
func (m *mockReservationRepo) Extend(ctx context.Context, id, userID uuid.UUID) (*entity.Reservation, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Reservation), args.Error(1)
}
func (m *mockReservationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *mockReservationRepo) CancelOverdue(ctx context.Context) ([]entity.Reservation, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Reservation), args.Error(1)
}
func (m *mockReservationRepo) FindApproachingDeadline(ctx context.Context, within time.Duration) ([]entity.Reservation, error) {
	args := m.Called(ctx, within)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Reservation), args.Error(1)
}
func (m *mockReservationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockReservationRepo) ActivateByQRToken(ctx context.Context, qrToken string, libraryID uuid.UUID) (*entity.Reservation, error) {
	args := m.Called(ctx, qrToken, libraryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Reservation), args.Error(1)
}
func (m *mockReservationRepo) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReservationView, error) {
	return nil, nil
}
func (m *mockReservationRepo) ListByUserView(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.ReservationView, int, error) {
	return nil, 0, nil
}
func (m *mockReservationRepo) ListByLibraryView(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.ReservationView, int, error) {
	return nil, 0, nil
}
func (m *mockReservationRepo) ListAllView(ctx context.Context, limit, offset int, status *string) ([]*entity.ReservationView, int, error) {
	return nil, 0, nil
}

type mockNotifRepo struct{ mock.Mock }

func (m *mockNotifRepo) Create(ctx context.Context, n *entity.Notification) (*entity.Notification, error) {
	args := m.Called(ctx, n)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Notification), args.Error(1)
}
func (m *mockNotifRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entity.Notification), args.Int(1), args.Error(2)
}
func (m *mockNotifRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockNotifRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return m.Called(ctx, id, userID).Error(0)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func newUC(resRepo *mockReservationRepo, notifRepo *mockNotifRepo) domainUseCase.ReservationUseCase {
	return reservation.NewReservationUseCase(resRepo, notifRepo, nil)
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	userID := uuid.New()
	libBookID := uuid.New()
	dueDate := time.Now().Add(7 * 24 * time.Hour)

	input := &entity.Reservation{
		UserID:         userID,
		LibraryBookID:  libBookID,
		DueDate:        dueDate,
	}
	created := &entity.Reservation{
		ID:             uuid.New(),
		UserID:         userID,
		LibraryBookID:  libBookID,
		Status:         entity.ReservationPending,
		ReservedAt:     time.Now(),
		DueDate:        dueDate,
	}
	resRepo.On("CreateWithDecrement", mock.Anything, mock.AnythingOfType("*entity.Reservation")).Return(created, nil)
	notifRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Notification")).Return((*entity.Notification)(nil), nil).Maybe()

	result, err := uc.Create(context.Background(), input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, entity.ReservationPending, result.Status)
	resRepo.AssertExpectations(t)
}

func TestCreate_RepoError(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	resRepo.On("CreateWithDecrement", mock.Anything, mock.AnythingOfType("*entity.Reservation")).
		Return(nil, errors.New("no available copies"))

	_, err := uc.Create(context.Background(), &entity.Reservation{
		UserID:        uuid.New(),
		LibraryBookID: uuid.New(),
		DueDate:       time.Now().Add(7 * 24 * time.Hour),
	})

	assert.Error(t, err)
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

func TestCancel_Success(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	id := uuid.New()
	callerID := uuid.New()
	resRepo.On("CancelWithIncrement", mock.Anything, id, callerID).Return(nil)
	notifRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Notification")).Return(nil, nil).Maybe()

	err := uc.Cancel(context.Background(), id, callerID)

	assert.NoError(t, err)
	resRepo.AssertExpectations(t)
}

func TestCancel_Forbidden(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	id := uuid.New()
	callerID := uuid.New()
	resRepo.On("CancelWithIncrement", mock.Anything, id, callerID).Return(entity.ErrForbidden)

	err := uc.Cancel(context.Background(), id, callerID)

	assert.ErrorIs(t, err, entity.ErrForbidden)
}

// ─── Extend ───────────────────────────────────────────────────────────────────

func TestExtend_Success(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	id := uuid.New()
	userID := uuid.New()
	newDue := time.Now().Add(14 * 24 * time.Hour)
	expected := &entity.Reservation{ID: id, UserID: userID, DueDate: newDue}

	resRepo.On("Extend", mock.Anything, id, userID).Return(expected, nil)
	notifRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Notification")).Return(nil, nil).Maybe()

	result, err := uc.Extend(context.Background(), id, userID)

	assert.NoError(t, err)
	assert.Equal(t, newDue, result.DueDate)
}

// ─── StaffUpdateStatus ────────────────────────────────────────────────────────

func TestStaffUpdateStatus_Success(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	id := uuid.New()
	libraryID := uuid.New()
	resRepo.On("StaffUpdateStatus", mock.Anything, id, libraryID, entity.ReservationActive).Return(nil)

	err := uc.StaffUpdateStatus(context.Background(), id, libraryID, entity.ReservationActive)

	assert.NoError(t, err)
	resRepo.AssertExpectations(t)
}

func TestStaffUpdateStatus_Forbidden(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	id := uuid.New()
	libraryID := uuid.New()
	resRepo.On("StaffUpdateStatus", mock.Anything, id, libraryID, entity.ReservationActive).Return(entity.ErrForbidden)

	err := uc.StaffUpdateStatus(context.Background(), id, libraryID, entity.ReservationActive)

	assert.ErrorIs(t, err, entity.ErrForbidden)
}

// ─── UpdateStatus (admin) ─────────────────────────────────────────────────────

func TestUpdateStatus_Success(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	id := uuid.New()
	resRepo.On("UpdateStatus", mock.Anything, id, entity.ReservationCompleted).Return(nil)

	err := uc.UpdateStatus(context.Background(), id, entity.ReservationCompleted)

	assert.NoError(t, err)
	resRepo.AssertExpectations(t)
}

// ─── ListByUser ───────────────────────────────────────────────────────────────

func TestListByUser_Success(t *testing.T) {
	resRepo := new(mockReservationRepo)
	notifRepo := new(mockNotifRepo)
	uc := newUC(resRepo, notifRepo)

	userID := uuid.New()
	items := []*entity.Reservation{{ID: uuid.New(), UserID: userID}}
	resRepo.On("ListByUser", mock.Anything, userID, 10, 0).Return(items, 1, nil)

	result, total, err := uc.ListByUser(context.Background(), userID, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, result, 1)
}
