package reading_session

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.ReadingSessionUseCase
}

func NewHandler(uc domainUseCase.ReadingSessionUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать/обновить сессию чтения
// @Tags        reading-sessions
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body upsertReadingSessionRequest true "Данные сессии"
// @Success     200 {object} readingSessionResponse
// @Router      /reading-sessions [post]
func (h *Handler) Upsert(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req upsertReadingSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	session := &entity.ReadingSession{
		UserID:      userID,
		BookID:      bookID,
		CurrentPage: req.CurrentPage,
		Status:      entity.ReadingStatus(req.Status),
	}

	result, err := h.uc.Upsert(c.Request.Context(), session)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toReadingSessionResponse(result))
}

// @Summary     Сессия по книге
// @Tags        reading-sessions
// @Security    BearerAuth
// @Produce     json
// @Param       book_id path string true "ID книги"
// @Success     200 {object} readingSessionResponse
// @Router      /reading-sessions/book/{book_id} [get]
func (h *Handler) GetByBook(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	session, err := h.uc.GetByUserAndBook(c.Request.Context(), userID, bookID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toReadingSessionResponse(session))
}

// @Summary     Мои сессии чтения
// @Tags        reading-sessions
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Router      /reading-sessions [get]
func (h *Handler) ListByUser(c *gin.Context) {
	userID := middleware.GetUserID(c)

	sessions, err := h.uc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	items := make([]*readingSessionResponse, len(sessions))
	for i, s := range sessions {
		resp := toReadingSessionResponse(s)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Удалить сессию
// @Tags        reading-sessions
// @Security    BearerAuth
// @Param       id path string true "ID сессии"
// @Success     204
// @Router      /reading-sessions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
func toReadingSessionResponse(s *entity.ReadingSession) readingSessionResponse {
	resp := readingSessionResponse{
		ID:          s.ID.String(),
		UserID:      s.UserID.String(),
		BookID:      s.BookID.String(),
		CurrentPage: s.CurrentPage,
		Status:      string(s.Status),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.FinishedAt != nil {
		t := s.FinishedAt.Format("2006-01-02T15:04:05Z")
		resp.FinishedAt = &t
	}
	return resp
}
