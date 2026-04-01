package book_file

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
)

type Handler struct {
	uc domainUseCase.BookFileUseCase
}

func NewHandler(uc domainUseCase.BookFileUseCase) *Handler {
	return &Handler{uc: uc}
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

// @Summary     Загрузить файл книги
// @Description Загружает файл в MinIO. Формат определяется автоматически по расширению файла (.pdf, .epub, .mp3).
//
//	Для PDF количество страниц считывается автоматически.
//	Лимиты размера: PDF/EPUB — 50 MB, MP3 — 200 MB.
//
// @Tags        book-files
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       book_id formData string true "ID книги"
// @Param       file    formData file   true "Файл (.pdf, .epub, .mp3)"
// @Success     201 {object} bookFileResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string "Файл уже загружен для этой книги"
// @Router      /admin/book-files/upload [post]
func (h *Handler) Upload(c *gin.Context) {
	bookID, err := uuid.Parse(c.PostForm("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	result, err := h.uc.Upload(c.Request.Context(), bookID, &fileupload.File{
		Filename:    fh.Filename,
		Reader:      f,
		Size:        fh.Size,
		ContentType: fh.Header.Get("Content-Type"),
	})
	if err != nil {
		if errors.Is(err, entity.ErrFileLimitExceeded) {
			c.JSON(http.StatusConflict, gin.H{"error": "a file of this type is already uploaded for this book"})
			return
		}
		if errors.Is(err, entity.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, toBookFileResponse(result))
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
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func toBookFileResponse(f *entity.BookFile) bookFileResponse {
	return bookFileResponse{
		ID:      f.ID.String(),
		BookID:  f.BookID.String(),
		Format:  string(f.Format),
		FileURL: f.FileURL,
	}
}
