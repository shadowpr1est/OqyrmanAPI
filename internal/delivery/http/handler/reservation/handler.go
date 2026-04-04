package reservation

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.ReservationUseCase
}

func NewHandler(uc domainUseCase.ReservationUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать бронь
// @Tags        reservations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createReservationRequest true "Данные брони"
// @Success     201 {object} reservationViewResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /reservations [post]
func (h *Handler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req createReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date format, use YYYY-MM-DD"})
		return
	}
	today := time.Now().Truncate(24 * time.Hour)
	if dueDate.Before(today) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date cannot be in the past"})
		return
	}
	if dueDate.After(today.AddDate(0, 0, 14)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date cannot be more than 14 days from today"})
		return
	}

	libraryBookID, err := uuid.Parse(req.LibraryBookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_book_id"})
		return
	}

	result, err := h.uc.Create(c.Request.Context(), &entity.Reservation{
		UserID:        userID,
		LibraryBookID: libraryBookID,
		DueDate:       dueDate,
	})
	if err != nil {
		if errors.Is(err, entity.ErrNoAvailableCopies) {
			c.JSON(http.StatusConflict, gin.H{"error": "no available copies"})
			return
		}
		if errors.Is(err, entity.ErrDuplicateReservation) {
			c.JSON(http.StatusConflict, gin.H{"error": "you already have an active reservation for this book"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	view, err := h.uc.GetByIDView(c.Request.Context(), result.ID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	c.JSON(http.StatusCreated, toReservationViewResponse(view, false))
}

// @Summary     Получить бронь
// @Tags        reservations
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID брони"
// @Success     200 {object} reservationViewResponse
// @Failure     404 {object} map[string]string
// @Router      /reservations/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	r, err := h.uc.GetByIDView(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrReservationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	// Only the owner, staff, and admin may view a reservation.
	callerID := middleware.GetUserID(c)
	role := middleware.GetRole(c)
	if r.UserID != callerID && role != "Admin" && role != "Staff" {
		common.Forbidden(c)
		return
	}

	c.JSON(http.StatusOK, toReservationViewResponse(r, role == "Admin" || role == "Staff"))
}

// @Summary     Мои брони
// @Tags        reservations
// @Security    BearerAuth
// @Produce     json
// @Param       limit  query int false "Лимит"    default(20)
// @Param       offset query int false "Смещение" default(0)
// @Success     200 {object} map[string]interface{}
// @Router      /reservations [get]
func (h *Handler) ListByUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	limit, offset, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, total, err := h.uc.ListByUserView(c.Request.Context(), userID, limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationViewResponses(items, false),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary     Отменить бронь
// @Tags        reservations
// @Security    BearerAuth
// @Param       id path string true "ID брони"
// @Success     204
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /reservations/{id}/cancel [patch]
func (h *Handler) Cancel(c *gin.Context) {
	callerID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Cancel(c.Request.Context(), id, callerID); err != nil {
		handleReservationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Продлить бронь
// @Tags        reservations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string              true "ID брони"
// @Param       input body extendReservationRequest true "Новая дата возврата"
// @Success     200 {object} reservationViewResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /reservations/{id}/extend [put]
func (h *Handler) Extend(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req extendReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	newDueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date format, use YYYY-MM-DD"})
		return
	}
	extendToday := time.Now().Truncate(24 * time.Hour)
	if newDueDate.Before(extendToday) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date cannot be in the past"})
		return
	}
	if newDueDate.After(extendToday.AddDate(0, 0, 14)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date cannot be more than 14 days from today"})
		return
	}

	result, err := h.uc.Extend(c.Request.Context(), id, userID, newDueDate)
	if err != nil {
		handleReservationError(c, err)
		return
	}

	view, err := h.uc.GetByIDView(c.Request.Context(), result.ID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, toReservationViewResponse(view, false))
}

// ─── Staff ────────────────────────────────────────────────────────────────────

// @Summary     Брони библиотеки
// @Tags        reservations
// @Security    BearerAuth
// @Produce     json
// @Param       limit  query int    false "Лимит"    default(20)
// @Param       offset query int    false "Смещение" default(0)
// @Param       status query string false "Фильтр по статусу"
// @Success     200 {object} map[string]interface{}
// @Failure     403 {object} map[string]string
// @Router      /staff/reservations [get]
func (h *Handler) ListByLibrary(c *gin.Context) {
	libraryID := middleware.GetLibraryID(c)
	if libraryID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no library assigned"})
		return
	}

	limit, offset, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var statusPtr *string
	if s := c.Query("status"); s != "" {
		statusPtr = &s
	}

	items, total, err := h.uc.ListByLibraryView(c.Request.Context(), *libraryID, limit, offset, statusPtr)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationViewResponses(items, true),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary     Отменить бронь (staff)
// @Tags        reservations
// @Security    BearerAuth
// @Param       id path string true "ID брони"
// @Success     204
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /staff/reservations/{id}/cancel [patch]
func (h *Handler) StaffCancel(c *gin.Context) {
	libraryID := middleware.GetLibraryID(c)
	if libraryID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no library assigned"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.StaffCancel(c.Request.Context(), id, *libraryID); err != nil {
		handleReservationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Возврат книги (staff)
// @Tags        reservations
// @Security    BearerAuth
// @Param       id path string true "ID брони"
// @Success     204
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /staff/reservations/{id}/return [patch]
func (h *Handler) StaffReturn(c *gin.Context) {
	libraryID := middleware.GetLibraryID(c)
	if libraryID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no library assigned"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.StaffReturn(c.Request.Context(), id, *libraryID); err != nil {
		handleReservationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ─── Admin ────────────────────────────────────────────────────────────────────
// @Summary     Все брони (admin)
// @Tags        reservations
// @Security    BearerAuth
// @Produce     json
// @Param       limit  query int    false "Лимит"    default(20)
// @Param       offset query int    false "Смещение" default(0)
// @Param       status query string false "Фильтр по статусу"
// @Success     200 {object} map[string]interface{}
// @Router      /admin/reservations [get]
func (h *Handler) ListAll(c *gin.Context) {
	limit, offset, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var statusPtr *string
	if s := c.Query("status"); s != "" {
		statusPtr = &s
	}

	items, total, err := h.uc.ListAllView(c.Request.Context(), limit, offset, statusPtr)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationViewResponses(items, true),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary     Возврат книги (admin)
// @Tags        reservations
// @Security    BearerAuth
// @Param       id path string true "ID брони"
// @Success     204
// @Failure     404 {object} map[string]string
// @Router      /admin/reservations/{id}/return [patch]
func (h *Handler) AdminReturn(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.AdminReturn(c.Request.Context(), id); err != nil {
		handleReservationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Обновить статус брони (staff)
// @Tags        reservations
// @Security    BearerAuth
// @Accept      json
// @Param       id    path string            true "ID брони"
// @Param       input body updateStatusRequest true "Новый статус"
// @Success     204
// @Router      /staff/reservations/{id}/status [patch]
func (h *Handler) StaffUpdateStatus(c *gin.Context) {
	libraryID := middleware.GetLibraryID(c)
	if libraryID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no library assigned"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	status := entity.ReservationStatus(req.Status)
	if !isValidReservationStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	if err := h.uc.StaffUpdateStatus(c.Request.Context(), id, *libraryID, status); err != nil {
		handleReservationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Обновить статус брони (admin)
// @Tags        reservations
// @Security    BearerAuth
// @Accept      json
// @Param       id    path string            true "ID брони"
// @Param       input body updateStatusRequest true "Новый статус"
// @Success     204
// @Router      /admin/reservations/{id}/status [patch]
func (h *Handler) UpdateStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	status := entity.ReservationStatus(req.Status)
	if !isValidReservationStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	if err := h.uc.UpdateStatus(c.Request.Context(), id, status); err != nil {
		handleReservationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func isValidReservationStatus(s entity.ReservationStatus) bool {
	switch s {
	case entity.ReservationPending, entity.ReservationActive,
		entity.ReservationCompleted, entity.ReservationCancelled:
		return true
	}
	return false
}

func handleReservationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, entity.ErrReservationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
	case errors.Is(err, entity.ErrInvalidStatusTransition):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
	}
}

// toReservationViewResponse собирает ответ брони.
// includeUser=true — для staff/admin контекста (чужие брони); false — для личного списка пользователя.
func toReservationViewResponse(v *entity.ReservationView, includeUser bool) reservationViewResponse {
	resp := reservationViewResponse{
		ID:            v.ID.String(),
		Status:        string(v.Status),
		ReservedAt:    v.ReservedAt.Format(time.RFC3339),
		DueDate:       v.DueDate.Format("2006-01-02"),
		LibraryBookID: v.LibraryBookID.String(),
		Book: common.ReservationBookRef{
			ID:       v.BookID.String(),
			Title:    v.BookTitle,
			CoverURL: v.BookCoverURL,
		},
		Library: common.LibraryRef{
			ID:   v.LibraryID.String(),
			Name: v.LibraryName,
		},
	}

	if v.ReturnedAt != nil {
		s := v.ReturnedAt.Format(time.RFC3339)
		resp.ReturnedAt = &s
	}

	if includeUser {
		resp.User = &common.UserRef{
			ID:      v.UserID.String(),
			Name:    v.UserName,
			Surname: v.UserSurname,
			Email:   v.UserEmail,
		}
	}

	return resp
}

func toReservationViewResponses(items []*entity.ReservationView, includeUser bool) []reservationViewResponse {
	resp := make([]reservationViewResponse, len(items))
	for i, v := range items {
		resp[i] = toReservationViewResponse(v, includeUser)
	}
	return resp
}
