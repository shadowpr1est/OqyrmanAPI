package reservation

import (
	"errors"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toReservationResponse(r))
}

func (h *Handler) ListByUser(c *gin.Context) {
	userID := middleware.GetUserID(c)
	limit, offset, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, total, err := h.uc.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationResponses(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

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

// ─── Staff ────────────────────────────────────────────────────────────────────

// ListByLibrary — staff видит все брони своей библиотеки
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationResponses(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// StaffCancel — staff отменяет бронь только своей библиотеки
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

// StaffReturn — staff отмечает возврат книги только своей библиотеки
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  toReservationResponses(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
