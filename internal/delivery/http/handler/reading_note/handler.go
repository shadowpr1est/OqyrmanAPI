package notes

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
	uc domainUseCase.ReadingNoteUseCase
}

func NewHandler(uc domainUseCase.ReadingNoteUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать заметку
// @Tags        notes
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createNoteRequest true "Данные заметки"
// @Success     201 {object} noteResponse
// @Router      /notes [post]
func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	note := &entity.ReadingNote{
		UserID:  userID,
		BookID:  bookID,
		Page:    req.Page,
		Content: req.Content,
	}

	result, err := h.uc.Create(c.Request.Context(), note)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, toNoteResponse(result))
}

// @Summary     Получить заметку
// @Tags        notes
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID заметки"
// @Success     200 {object} noteResponse
// @Router      /notes/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	note, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if note.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	c.JSON(http.StatusOK, toNoteResponse(note))
}

// @Summary     Заметки по книге
// @Tags        notes
// @Security    BearerAuth
// @Produce     json
// @Param       book_id path string true "ID книги"
// @Success     200 {object} map[string]interface{}
// @Router      /notes/book/{book_id} [get]
func (h *Handler) ListByBook(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	notes, err := h.uc.ListByUserAndBook(c.Request.Context(), userID, bookID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	items := make([]*noteResponse, len(notes))
	for i, n := range notes {
		resp := toNoteResponse(n)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Обновить заметку
// @Tags        notes
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string          true "ID заметки"
// @Param       input body updateNoteRequest true "Данные для обновления"
// @Success     200 {object} noteResponse
// @Router      /notes/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	if req.Page != nil {
		existing.Page = *req.Page
	}
	if req.Content != nil {
		existing.Content = *req.Content
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toNoteResponse(result))
}

// @Summary     Удалить заметку
// @Tags        notes
// @Security    BearerAuth
// @Param       id path string true "ID заметки"
// @Success     204
// @Router      /notes/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
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

func toNoteResponse(n *entity.ReadingNote) noteResponse {
	return noteResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		BookID:    n.BookID.String(),
		Page:      n.Page,
		Content:   n.Content,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
