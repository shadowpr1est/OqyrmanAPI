package reservation

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
// @Success     201 {object} reservationResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /reservations [post]
func (h *Handler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req createReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date format, use YYYY-MM-DD"})
		return
	}
	if dueDate.Before(time.Now().Truncate(24 * time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date cannot be in the past"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toReservationResponse(result))
}

// @Summary     Получить бронь
// @Tags        reservations
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID брони"
// @Success     200 {object} reservationResponse
// @Failure     404 {object} map[string]string
// @Router      /reservations/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	r, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrReservationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "reservation not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toReservationResponse(r))
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

	items, total, err := h.uc.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationResponses(items),
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
// @Success     200 {object} reservationResponse
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newDueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date format, use YYYY-MM-DD"})
		return
	}
	if newDueDate.Before(time.Now().Truncate(24 * time.Hour)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "due_date cannot be in the past"})
		return
	}

	result, err := h.uc.Extend(c.Request.Context(), id, userID, newDueDate)
	if err != nil {
		handleReservationError(c, err)
		return
	}

	c.JSON(http.StatusOK, toReservationResponse(result))
}

// ─── Staff ────────────────────────────────────────────────────────────────────

// @Summary     Брони библиотеки
// @Tags        staff
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

	items, total, err := h.uc.ListByLibrary(c.Request.Context(), *libraryID, limit, offset, statusPtr)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationResponses(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary     Отменить бронь (staff)
// @Tags        staff
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
// @Tags        staff
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
// @Tags        admin
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

	items, total, err := h.uc.ListAll(c.Request.Context(), limit, offset, statusPtr)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationResponses(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary     Возврат книги (admin)
// @Tags        admin
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
// @Tags        staff
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.StaffUpdateStatus(
		c.Request.Context(), id, *libraryID,
		entity.ReservationStatus(req.Status),
	); err != nil {
		handleReservationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Обновить статус брони (admin)
// @Tags        admin
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.UpdateStatus(c.Request.Context(), id, entity.ReservationStatus(req.Status)); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func toReservationResponse(r *entity.Reservation) reservationResponse {
	resp := reservationResponse{
		ID:            r.ID.String(),
		UserID:        r.UserID.String(),
		LibraryBookID: r.LibraryBookID.String(),
		Status:        string(r.Status),
		ReservedAt:    r.ReservedAt.Format("2006-01-02T15:04:05Z"),
		DueDate:       r.DueDate.Format("2006-01-02"),
	}
	if r.ReturnedAt != nil {
		s := r.ReturnedAt.Format("2006-01-02T15:04:05Z")
		resp.ReturnedAt = &s
	}
	return resp
}

func toReservationResponses(items []*entity.Reservation) []*reservationResponse {
	resp := make([]*reservationResponse, len(items))
	for i, r := range items {
		res := toReservationResponse(r)
		resp[i] = &res
	}
	return resp
}
