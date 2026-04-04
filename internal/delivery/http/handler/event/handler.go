package event

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
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
		common.InternalError(c)
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
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, toEventResponse(e))
}

// @Summary     Создать событие
// @Tags        events
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       title       formData string true  "Название"
// @Param       description formData string false "Описание"
// @Param       location    formData string false "Место"
// @Param       starts_at   formData string true  "Начало (RFC3339)"
// @Param       ends_at     formData string true  "Конец (RFC3339)"
// @Param       cover       formData file   false "Обложка (jpg, png, webp)"
// @Success     201 {object} eventResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/events [post]
func (h *Handler) Create(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
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

	cover, err := parseCover(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if cover != nil {
		defer cover.Reader.Close()
	}

	result, err := h.uc.Create(c.Request.Context(), &entity.Event{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
	}, cover)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toEventResponse(result))
}

// @Summary     Обновить событие
// @Tags        events
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       id          path     string true  "ID события"
// @Param       title       formData string true  "Название"
// @Param       description formData string false "Описание"
// @Param       location    formData string false "Место"
// @Param       starts_at   formData string true  "Начало (RFC3339)"
// @Param       ends_at     formData string true  "Конец (RFC3339)"
// @Param       cover       formData file   false "Обложка (jpg, png, webp)"
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
	if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
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

	cover, err := parseCover(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if cover != nil {
		defer cover.Reader.Close()
	}

	result, err := h.uc.Update(c.Request.Context(), &entity.Event{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartsAt:    startsAt,
		EndsAt:      endsAt,
	}, cover)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, toEventResponse(result))
}

// parseCover извлекает файл обложки из multipart-формы.
// Возвращает nil, nil если поле cover не передано.
func parseCover(c *gin.Context) (*fileupload.File, error) {
	fh, err := c.FormFile("cover")
	if err != nil {
		return nil, nil // обложка не обязательна
	}
	f, err := fh.Open()
	if err != nil {
		return nil, errors.New("cannot open cover file")
	}

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	ct := http.DetectContentType(buf[:n])
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		f.Close()
		return nil, errors.New("cannot process cover file")
	}
	switch ct {
	case "image/jpeg", "image/png", "image/webp":
		// ok
	default:
		f.Close()
		return nil, errors.New("only jpeg, png, webp images are allowed")
	}

	return &fileupload.File{
		Filename:    fh.Filename,
		Reader:      f,
		Size:        fh.Size,
		ContentType: ct,
	}, nil
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
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}
