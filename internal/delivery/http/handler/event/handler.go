package event

import (
	"log/slog"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.EventUseCase
}

func NewHandler(uc domainUseCase.EventUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Список событий
// @Tags        events
// @Produce     json
// @Param       limit  query int false "Лимит (default 20)"
// @Param       offset query int false "Смещение (default 0)"
// @Success     200 {object} map[string]interface{}
// @Router      /events [get]
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := h.uc.List(c.Request.Context(), limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]eventResponse, len(items))
	for i, e := range items {
		resp[i] = toEventResponse(e)
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  resp,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary     Детали события
// @Tags        events
// @Produce     json
// @Param       id path string true "ID события"
// @Success     200 {object} eventResponse
// @Router      /events/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	e, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toEventResponse(e))
}

// @Summary     Создать событие
// @Tags        events
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createEventRequest true "Данные события"
// @Success     201 {object} eventResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/events [post]
func (h *Handler) Create(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid starts_at, use RFC3339"})
		return
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ends_at, use RFC3339"})
		return
	}
	if !endsAt.After(startsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ends_at must be after starts_at"})
		return
	}

	result, err := h.uc.Create(c.Request.Context(), &entity.Event{
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		Location:    req.Location,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
	})
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, toEventResponse(result))
}

// @Summary     Обновить событие
// @Tags        events
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string           true "ID события"
// @Param       input body updateEventRequest true "Данные события"
// @Success     200 {object} eventResponse
// @Failure     400 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /admin/events/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid starts_at, use RFC3339"})
		return
	}
	endsAt, err := time.Parse(time.RFC3339, req.EndsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ends_at, use RFC3339"})
		return
	}
	if !endsAt.After(startsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ends_at must be after starts_at"})
		return
	}

	result, err := h.uc.Update(c.Request.Context(), &entity.Event{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		Location:    req.Location,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
	})
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toEventResponse(result))
}

// @Summary     Удалить событие
// @Tags        events
// @Security    BearerAuth
// @Param       id path string true "ID события"
// @Success     204
// @Failure     404 {object} map[string]string
// @Router      /admin/events/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
