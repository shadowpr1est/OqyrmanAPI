package reservation

import (
	"net/http"
	"strconv"
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
// @Router      /reservations [post]
func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

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

	r := &entity.Reservation{
		UserID:     userID,
		SourceType: entity.SourceType(req.SourceType),
		DueDate:    dueDate,
	}

	if req.LibraryBookID != nil {
		id, err := uuid.Parse(*req.LibraryBookID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_book_id"})
			return
		}
		r.LibraryBookID = &id
	}

	if req.MachineBookID != nil {
		id, err := uuid.Parse(*req.MachineBookID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid machine_book_id"})
			return
		}
		r.MachineBookID = &id
	}

	result, err := h.uc.Create(c.Request.Context(), r)
	if err != nil {
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
// @Router      /reservations/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	r, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
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
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.uc.ListByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]*reservationResponse, len(items))
	for i, r := range items {
		res := toReservationResponse(r)
		resp[i] = &res
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  resp,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary     Обновить статус брони
// @Tags        reservations
// @Security    BearerAuth
// @Accept      json
// @Produce     json
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Отменить бронь
// @Tags        reservations
// @Security    BearerAuth
// @Param       id path string true "ID брони"
// @Success     204
// @Router      /reservations/{id}/cancel [patch]
func (h *Handler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Cancel(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Отметить книгу как возвращённую
// @Tags        reservations
// @Security    BearerAuth
// @Param       id path string true "ID брони"
// @Success     204
// @Router      /admin/reservations/{id}/return [patch]
func (h *Handler) Return(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Return(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Все брони (admin)
// @Tags        reservations
// @Security    BearerAuth
// @Produce     json
// @Param       limit  query int    false "Лимит"            default(20)
// @Param       offset query int    false "Смещение"         default(0)
// @Param       status query string false "Фильтр по статусу (pending, active, completed, cancelled)"
// @Success     200 {object} map[string]interface{}
// @Router      /admin/reservations [get]
func (h *Handler) ListAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var statusPtr *string
	if status := c.Query("status"); status != "" {
		statusPtr = &status
	}

	items, total, err := h.uc.ListAll(c.Request.Context(), limit, offset, statusPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]*reservationResponse, len(items))
	for i, r := range items {
		res := toReservationResponse(r)
		resp[i] = &res
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  resp,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func toReservationResponse(r *entity.Reservation) reservationResponse {
	resp := reservationResponse{
		ID:         r.ID.String(),
		UserID:     r.UserID.String(),
		SourceType: string(r.SourceType),
		Status:     string(r.Status),
		ReservedAt: r.ReservedAt.Format("2006-01-02T15:04:05Z"),
		DueDate:    r.DueDate.Format("2006-01-02"),
	}

	if r.LibraryBookID != nil {
		s := r.LibraryBookID.String()
		resp.LibraryBookID = &s
	}
	if r.MachineBookID != nil {
		s := r.MachineBookID.String()
		resp.MachineBookID = &s
	}
	if r.ReturnedAt != nil {
		s := r.ReturnedAt.Format("2006-01-02T15:04:05Z")
		resp.ReturnedAt = &s
	}

	return resp
}
