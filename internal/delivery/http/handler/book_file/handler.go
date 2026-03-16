package book_file

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.BookFileUseCase
}

func NewHandler(uc domainUseCase.BookFileUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Добавить файл книги
// @Tags        book-files
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createBookFileRequest true "Данные файла"
// @Success     201 {object} bookFileResponse
// @Router      /admin/book-files [post]
func (h *Handler) Create(c *gin.Context) {
	var req createBookFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	file := &entity.BookFile{
		BookID:  bookID,
		Format:  req.Format,
		FileURL: req.FileURL,
		IsAudio: req.IsAudio,
	}

	result, err := h.uc.Create(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toBookFileResponse(result))
}

// @Summary     Получить файл
// @Tags        book-files
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID файла"
// @Success     200 {object} bookFileResponse
// @Router      /book-files/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	file, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toBookFileResponse(file))
}

// @Summary     Файлы книги
// @Tags        book-files
// @Security    BearerAuth
// @Produce     json
// @Param       book_id path string true "ID книги"
// @Success     200 {object} map[string]interface{}
// @Router      /book-files/book/{book_id} [get]
func (h *Handler) ListByBook(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	files, err := h.uc.ListByBook(c.Request.Context(), bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookFileResponse, len(files))
	for i, f := range files {
		resp := toBookFileResponse(f)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Удалить файл
// @Tags        book-files
// @Security    BearerAuth
// @Param       id path string true "ID файла"
// @Success     204
// @Router      /admin/book-files/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func toBookFileResponse(f *entity.BookFile) bookFileResponse {
	return bookFileResponse{
		ID:      f.ID.String(),
		BookID:  f.BookID.String(),
		Format:  f.Format,
		FileURL: f.FileURL,
		IsAudio: f.IsAudio,
	}
}
