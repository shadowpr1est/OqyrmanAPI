package notes

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
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
	userID := middleware.GetUserID(c)

	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	note := &entity.ReadingNote{
		UserID:   userID,
		BookID:   bookID,
		Position: req.Position,
		Content:  req.Content,
	}

	result, err := h.uc.Create(c.Request.Context(), note)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toNoteResponse(result))
}

// @Summary     Получить заметку
// @Tags        notes
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID заметки"
// @Success     200 {object} noteViewResponse
// @Router      /notes/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	note, err := h.uc.GetByIDView(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	// Возвращаем 404 вместо 403, чтобы не раскрывать факт существования чужой записи.
	if note.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}
	c.JSON(http.StatusOK, toNoteViewResponse(note))
}

// @Summary     Заметки по книге
// @Tags        notes
// @Security    BearerAuth
// @Produce     json
// @Param       book_id path string true "ID книги"
// @Success     200 {object} map[string]interface{}
// @Router      /notes/book/{book_id} [get]
func (h *Handler) ListByBook(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	notes, err := h.uc.ListByUserAndBookView(c.Request.Context(), userID, bookID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*noteViewResponse, len(notes))
	for i, n := range notes {
		resp := toNoteViewResponse(n)
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
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "note not found")
		return
	}
	if existing.UserID != userID {
		common.Forbidden(c)
		return
	}

	if req.Position != nil {
		existing.Position = *req.Position
	}
	if req.Content != nil {
		existing.Content = *req.Content
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "note not found")
		return
	}
	if existing.UserID != userID {
		common.Forbidden(c)
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func toNoteResponse(n *entity.ReadingNote) noteResponse {
	return noteResponse{
		ID:        n.ID.String(),
		BookID:    n.BookID.String(),
		Position:  n.Position,
		Content:   n.Content,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: n.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toNoteViewResponse(v *entity.ReadingNoteView) noteViewResponse {
	return noteViewResponse{
		ID:        v.ID.String(),
		BookID:    v.BookID.String(),
		BookTitle: v.BookTitle,
		Position:  v.Position,
		Content:   v.Content,
		CreatedAt: v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
