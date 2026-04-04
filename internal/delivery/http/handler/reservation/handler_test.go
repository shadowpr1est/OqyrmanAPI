package reservation

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ─────────────────────────────────────────────────────────────────────

type mockReservationUseCase struct{ mock.Mock }

func (m *mockReservationUseCase) Create(ctx context.Context, r *entity.Reservation) (*entity.Reservation, error) {
	args := m.Called(ctx, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Reservation), args.Error(1)
}
func (m *mockReservationUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Reservation), args.Error(1)
}
func (m *mockReservationUseCase) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entity.Reservation), args.Int(1), args.Error(2)
}
func (m *mockReservationUseCase) Cancel(ctx context.Context, id, callerID uuid.UUID) error {
	return m.Called(ctx, id, callerID).Error(0)
}
func (m *mockReservationUseCase) Extend(ctx context.Context, id, userID uuid.UUID, newDueDate time.Time) (*entity.Reservation, error) {
	args := m.Called(ctx, id, userID, newDueDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Reservation), args.Error(1)
}
func (m *mockReservationUseCase) ListByLibrary(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	args := m.Called(ctx, libraryID, limit, offset, status)
	return args.Get(0).([]*entity.Reservation), args.Int(1), args.Error(2)
}
func (m *mockReservationUseCase) StaffCancel(ctx context.Context, id, libraryID uuid.UUID) error {
	return m.Called(ctx, id, libraryID).Error(0)
}
func (m *mockReservationUseCase) StaffReturn(ctx context.Context, id, libraryID uuid.UUID) error {
	return m.Called(ctx, id, libraryID).Error(0)
}
func (m *mockReservationUseCase) StaffUpdateStatus(ctx context.Context, id, libraryID uuid.UUID, status entity.ReservationStatus) error {
	return m.Called(ctx, id, libraryID, status).Error(0)
}
func (m *mockReservationUseCase) ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	args := m.Called(ctx, limit, offset, status)
	return args.Get(0).([]*entity.Reservation), args.Int(1), args.Error(2)
}
func (m *mockReservationUseCase) AdminReturn(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockReservationUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *mockReservationUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockReservationUseCase) CancelOverdue(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}
func (m *mockReservationUseCase) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReservationView, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ReservationView), args.Error(1)
}
func (m *mockReservationUseCase) ListByUserView(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.ReservationView, int, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entity.ReservationView), args.Int(1), args.Error(2)
}
func (m *mockReservationUseCase) ListByLibraryView(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.ReservationView, int, error) {
	args := m.Called(ctx, libraryID, limit, offset, status)
	return args.Get(0).([]*entity.ReservationView), args.Int(1), args.Error(2)
}
func (m *mockReservationUseCase) ListAllView(ctx context.Context, limit, offset int, status *string) ([]*entity.ReservationView, int, error) {
	args := m.Called(ctx, limit, offset, status)
	return args.Get(0).([]*entity.ReservationView), args.Int(1), args.Error(2)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

var (
	testUserID    = uuid.New()
	testLibraryID = uuid.New()
)

func init() {
	gin.SetMode(gin.TestMode)
}

func serveJSON(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// injectUser wraps a handler by injecting userID into the gin context.
func injectUser(userID uuid.UUID, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		h(c)
	}
}

// injectUserAndLibrary injects both userID and libraryID into the context.
func injectUserAndLibrary(userID uuid.UUID, libID *uuid.UUID, h gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Set("libraryID", libID)
		h(c)
	}
}

func makeReservation(id, userID uuid.UUID) *entity.Reservation {
	return &entity.Reservation{
		ID:            id,
		UserID:        userID,
		LibraryBookID: uuid.New(),
		Status:        entity.ReservationPending,
		ReservedAt:    time.Now(),
		DueDate:       time.Now().Add(7 * 24 * time.Hour),
	}
}

func makeReservationView(id, userID uuid.UUID) *entity.ReservationView {
	return &entity.ReservationView{
		ID:           id,
		Status:       entity.ReservationPending,
		ReservedAt:   time.Now(),
		DueDate:      time.Now().Add(7 * 24 * time.Hour),
		UserID:       userID,
		UserName:    "Test",
		UserSurname: "User",
		UserEmail:    "test@example.com",
		LibraryBookID: uuid.New(),
		BookID:       uuid.New(),
		BookTitle:    "Test Book",
		LibraryID:    uuid.New(),
		LibraryName:  "Test Library",
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.POST("/reservations", injectUser(testUserID, h.Create))

	libBookID := uuid.New()
	result := makeReservation(uuid.New(), testUserID)
	result.LibraryBookID = libBookID
	view := makeReservationView(result.ID, testUserID)

	uc.On("Create", mock.Anything, mock.AnythingOfType("*entity.Reservation")).Return(result, nil)
	uc.On("GetByIDView", mock.Anything, result.ID).Return(view, nil)

	validDate := time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")
	body := fmt.Sprintf(`{"library_book_id":"%s","due_date":"%s"}`, libBookID, validDate)
	w := serveJSON(r, http.MethodPost, "/reservations", body)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"id"`)
	uc.AssertExpectations(t)
}

func TestCreate_InvalidBody(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.POST("/reservations", injectUser(testUserID, h.Create))

	w := serveJSON(r, http.MethodPost, "/reservations", `{}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Create")
}

func TestCreate_BadDateFormat(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.POST("/reservations", injectUser(testUserID, h.Create))

	body := fmt.Sprintf(`{"library_book_id":"%s","due_date":"01-01-2027"}`, uuid.New())
	w := serveJSON(r, http.MethodPost, "/reservations", body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "YYYY-MM-DD")
}

func TestCreate_DateInPast(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.POST("/reservations", injectUser(testUserID, h.Create))

	body := fmt.Sprintf(`{"library_book_id":"%s","due_date":"2020-01-01"}`, uuid.New())
	w := serveJSON(r, http.MethodPost, "/reservations", body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "past")
}

func TestCreate_InvalidLibraryBookID(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.POST("/reservations", injectUser(testUserID, h.Create))

	validDate := time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")
	body := fmt.Sprintf(`{"library_book_id":"not-a-uuid","due_date":"%s"}`, validDate)
	w := serveJSON(r, http.MethodPost, "/reservations", body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid library_book_id")
}

func TestCreate_NoAvailableCopies(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.POST("/reservations", injectUser(testUserID, h.Create))

	uc.On("Create", mock.Anything, mock.AnythingOfType("*entity.Reservation")).
		Return(nil, entity.ErrNoAvailableCopies)

	validDate := time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")
	body := fmt.Sprintf(`{"library_book_id":"%s","due_date":"%s"}`, uuid.New(), validDate)
	w := serveJSON(r, http.MethodPost, "/reservations", body)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "no available copies")
}

func TestCreate_DuplicateReservation(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.POST("/reservations", injectUser(testUserID, h.Create))

	uc.On("Create", mock.Anything, mock.AnythingOfType("*entity.Reservation")).
		Return(nil, entity.ErrDuplicateReservation)

	validDate := time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02")
	body := fmt.Sprintf(`{"library_book_id":"%s","due_date":"%s"}`, uuid.New(), validDate)
	w := serveJSON(r, http.MethodPost, "/reservations", body)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "already have an active reservation")
}

// ─── GetByID ──────────────────────────────────────────────────────────────────

func TestGetByID_Success(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/reservations/:id", injectUser(testUserID, h.GetByID))

	id := uuid.New()
	view := makeReservationView(id, testUserID)
	uc.On("GetByIDView", mock.Anything, id).Return(view, nil)

	w := serveJSON(r, http.MethodGet, "/reservations/"+id.String(), "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), id.String())
}

func TestGetByID_InvalidUUID(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/reservations/:id", h.GetByID)

	w := serveJSON(r, http.MethodGet, "/reservations/not-a-uuid", "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "GetByIDView")
}

func TestGetByID_NotFound(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/reservations/:id", h.GetByID)

	id := uuid.New()
	uc.On("GetByIDView", mock.Anything, id).Return(nil, entity.ErrReservationNotFound)

	w := serveJSON(r, http.MethodGet, "/reservations/"+id.String(), "")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetByID_InternalError(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/reservations/:id", h.GetByID)

	id := uuid.New()
	uc.On("GetByIDView", mock.Anything, id).Return(nil, errors.New("db timeout"))

	w := serveJSON(r, http.MethodGet, "/reservations/"+id.String(), "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
}

// ─── ListByUser ───────────────────────────────────────────────────────────────

func TestListByUser_Success(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/reservations", injectUser(testUserID, h.ListByUser))

	items := []*entity.ReservationView{makeReservationView(uuid.New(), testUserID)}
	uc.On("ListByUserView", mock.Anything, testUserID, 20, 0).Return(items, 1, nil)

	w := serveJSON(r, http.MethodGet, "/reservations", "")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total"`)
}

func TestListByUser_InvalidPagination(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/reservations", injectUser(testUserID, h.ListByUser))

	req := httptest.NewRequest(http.MethodGet, "/reservations?limit=-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "ListByUserView")
}

func TestListByUser_InternalError(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/reservations", injectUser(testUserID, h.ListByUser))

	uc.On("ListByUserView", mock.Anything, testUserID, 20, 0).
		Return([]*entity.ReservationView{}, 0, errors.New("connection pool exhausted"))

	w := serveJSON(r, http.MethodGet, "/reservations", "")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

func TestCancel_Success(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.PATCH("/reservations/:id/cancel", injectUser(testUserID, h.Cancel))

	id := uuid.New()
	uc.On("Cancel", mock.Anything, id, testUserID).Return(nil)

	req := httptest.NewRequest(http.MethodPatch, "/reservations/"+id.String()+"/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCancel_InvalidUUID(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.PATCH("/reservations/:id/cancel", injectUser(testUserID, h.Cancel))

	req := httptest.NewRequest(http.MethodPatch, "/reservations/bad-id/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancel_Forbidden(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.PATCH("/reservations/:id/cancel", injectUser(testUserID, h.Cancel))

	id := uuid.New()
	uc.On("Cancel", mock.Anything, id, testUserID).Return(entity.ErrForbidden)

	req := httptest.NewRequest(http.MethodPatch, "/reservations/"+id.String()+"/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCancel_NotFound(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.PATCH("/reservations/:id/cancel", injectUser(testUserID, h.Cancel))

	id := uuid.New()
	uc.On("Cancel", mock.Anything, id, testUserID).Return(entity.ErrReservationNotFound)

	req := httptest.NewRequest(http.MethodPatch, "/reservations/"+id.String()+"/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ─── ListByLibrary ────────────────────────────────────────────────────────────

func TestListByLibrary_NoLibraryID(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/staff/reservations", injectUserAndLibrary(testUserID, nil, h.ListByLibrary))

	req := httptest.NewRequest(http.MethodGet, "/staff/reservations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "no library assigned")
}

func TestListByLibrary_Success(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/staff/reservations", injectUserAndLibrary(testUserID, &testLibraryID, h.ListByLibrary))

	items := []*entity.ReservationView{makeReservationView(uuid.New(), testUserID)}
	uc.On("ListByLibraryView", mock.Anything, testLibraryID, 20, 0, (*string)(nil)).
		Return(items, 1, nil)

	req := httptest.NewRequest(http.MethodGet, "/staff/reservations", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"total"`)
}

func TestListByLibrary_WithStatusFilter(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.GET("/staff/reservations", injectUserAndLibrary(testUserID, &testLibraryID, h.ListByLibrary))

	status := "active"
	uc.On("ListByLibraryView", mock.Anything, testLibraryID, 20, 0, &status).
		Return([]*entity.ReservationView{}, 0, nil)

	req := httptest.NewRequest(http.MethodGet, "/staff/reservations?status=active", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─── StaffCancel ──────────────────────────────────────────────────────────────

func TestStaffCancel_Success(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.PATCH("/staff/reservations/:id/cancel", injectUserAndLibrary(testUserID, &testLibraryID, h.StaffCancel))

	id := uuid.New()
	uc.On("StaffCancel", mock.Anything, id, testLibraryID).Return(nil)

	req := httptest.NewRequest(http.MethodPatch, "/staff/reservations/"+id.String()+"/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestStaffCancel_NotFound(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.PATCH("/staff/reservations/:id/cancel", injectUserAndLibrary(testUserID, &testLibraryID, h.StaffCancel))

	id := uuid.New()
	uc.On("StaffCancel", mock.Anything, id, testLibraryID).Return(entity.ErrReservationNotFound)

	req := httptest.NewRequest(http.MethodPatch, "/staff/reservations/"+id.String()+"/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStaffCancel_InvalidStatusTransition(t *testing.T) {
	uc := new(mockReservationUseCase)
	h := NewHandler(uc)

	r := gin.New()
	r.PATCH("/staff/reservations/:id/cancel", injectUserAndLibrary(testUserID, &testLibraryID, h.StaffCancel))

	id := uuid.New()
	uc.On("StaffCancel", mock.Anything, id, testLibraryID).Return(entity.ErrInvalidStatusTransition)

	req := httptest.NewRequest(http.MethodPatch, "/staff/reservations/"+id.String()+"/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}
