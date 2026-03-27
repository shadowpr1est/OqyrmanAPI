package book

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
)

type Handler struct {
	uc        domainUseCase.BookUseCase
	libBookUC domainUseCase.LibraryBookUseCase
}

func NewHandler(uc domainUseCase.BookUseCase, libBookUC domainUseCase.LibraryBookUseCase) *Handler {
	return &Handler{uc: uc, libBookUC: libBookUC}
}

// @Summary     Создать книгу
// @Tags        books
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       author_id   formData string true  "ID автора"
// @Param       genre_id    formData string true  "ID жанра"
// @Param       title       formData string true  "Название"
// @Param       isbn        formData string false "ISBN"
// @Param       description formData string false "Описание"
// @Param       language    formData string false "Язык"
// @Param       year        formData int    false "Год"
// @Param       cover       formData file   false "Обложка (jpg, png)"
// @Success     201 {object} bookResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/books [post]
func (h *Handler) Create(c *gin.Context) {
	var req createBookRequest
	if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author_id"})
		return
	}
	genreID, err := uuid.Parse(req.GenreID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre_id"})
		return
	}

	book := &entity.Book{
		AuthorID:    authorID,
		GenreID:     genreID,
		Title:       req.Title,
		ISBN:        req.ISBN,
		Description: req.Description,
		Language:    req.Language,
		Year:        req.Year,
		TotalPages:  req.TotalPages,
	}

	var cover *fileupload.File
	if fh, err := c.FormFile("cover"); err == nil {
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open cover file"})
			return
		}
		defer f.Close()
		contentType := fh.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "image/jpeg"
		}
		cover = &fileupload.File{
			Filename:    fh.Filename,
			Reader:      f,
			Size:        fh.Size,
			ContentType: contentType,
		}
	}

	result, err := h.uc.Create(c.Request.Context(), book, cover)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toBookResponse(result))
}

// @Summary     Загрузить обложку книги
// @Tags        books
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       id    path     string true "ID книги"
// @Param       cover formData file   true "Обложка (jpg, png)"
// @Success     200 {object} bookResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/books/{id}/cover [post]
func (h *Handler) UploadCover(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	fh, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cover file is required"})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open cover file"})
		return
	}
	defer f.Close()

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	book, err := h.uc.UploadCover(c.Request.Context(), id, &fileupload.File{
		Filename:    fh.Filename,
		Reader:      f,
		Size:        fh.Size,
		ContentType: contentType,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toBookResponse(book))
}

// @Summary     Получить книгу
// @Tags        books
// @Produce     json
// @Param       id path string true "ID книги"
// @Success     200 {object} bookResponse
// @Failure     404 {object} map[string]string
// @Router      /books/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toBookResponse(book))
}

// @Summary     Список книг
// @Tags        books
// @Produce     json
// @Param       limit  query int false "Лимит"  default(20)
// @Param       offset query int false "Отступ" default(0)
// @Success     200 {object} listBookResponse
// @Router      /books [get]
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	books, total, err := h.uc.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookResponse, len(books))
	for i, b := range books {
		resp := toBookResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listBookResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Книги по автору
// @Tags        books
// @Produce     json
// @Param       author_id path string true "ID автора"
// @Success     200 {object} map[string]interface{}
// @Router      /books/author/{author_id} [get]
func (h *Handler) ListByAuthor(c *gin.Context) {
	authorID, err := uuid.Parse(c.Param("author_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author_id"})
		return
	}

	books, err := h.uc.ListByAuthor(c.Request.Context(), authorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookResponse, len(books))
	for i, b := range books {
		resp := toBookResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Книги по жанру
// @Tags        books
// @Produce     json
// @Param       genre_id path string true "ID жанра"
// @Success     200 {object} map[string]interface{}
// @Router      /books/genre/{genre_id} [get]
func (h *Handler) ListByGenre(c *gin.Context) {
	genreID, err := uuid.Parse(c.Param("genre_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre_id"})
		return
	}

	books, err := h.uc.ListByGenre(c.Request.Context(), genreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookResponse, len(books))
	for i, b := range books {
		resp := toBookResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Поиск книг
// @Tags        books
// @Produce     json
// @Param       q      query string true  "Поисковый запрос"
// @Param       limit  query int    false "Лимит"    default(20)
// @Param       offset query int    false "Смещение" default(0)
// @Success     200 {object} listBookResponse
// @Failure     400 {object} map[string]string
// @Router      /books/search [get]
func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query param q is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	books, total, err := h.uc.Search(c.Request.Context(), query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookResponse, len(books))
	for i, b := range books {
		resp := toBookResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listBookResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Обновить книгу
// @Tags        books
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string          true "ID книги"
// @Param       input body updateBookRequest true "Данные для обновления"
// @Success     200 {object} bookResponse
// @Failure     404 {object} map[string]string
// @Router      /admin/books/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.AuthorID != nil {
		authorID, err := uuid.Parse(*req.AuthorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author_id"})
			return
		}
		existing.AuthorID = authorID
	}
	if req.GenreID != nil {
		genreID, err := uuid.Parse(*req.GenreID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre_id"})
			return
		}
		existing.GenreID = genreID
	}
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.ISBN != nil {
		existing.ISBN = *req.ISBN
	}
	if req.CoverURL != nil {
		existing.CoverURL = *req.CoverURL
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Language != nil {
		existing.Language = *req.Language
	}
	if req.Year != nil {
		existing.Year = *req.Year
	}
	if req.TotalPages != nil {
		existing.TotalPages = req.TotalPages
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toBookResponse(result))
}

// @Summary     Удалить книгу
// @Tags        books
// @Security    BearerAuth
// @Param       id path string true "ID книги"
// @Success     204
// @Router      /admin/books/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Популярные книги
// @Tags        books
// @Produce     json
// @Param       limit  query int false "Лимит"  default(20)
// @Param       offset query int false "Отступ" default(0)
// @Success     200 {object} listBookResponse
// @Router      /books/popular [get]
func (h *Handler) ListPopular(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	books, total, err := h.uc.ListPopular(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookResponse, len(books))
	for i, b := range books {
		resp := toBookResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listBookResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Похожие книги
// @Tags        books
// @Produce     json
// @Param       id    path string true  "ID книги"
// @Param       limit query int   false "Лимит" default(10)
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Router      /books/{id}/similar [get]
func (h *Handler) ListSimilar(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	books, err := h.uc.ListSimilar(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookResponse, len(books))
	for i, b := range books {
		resp := toBookResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func toBookResponse(b *entity.Book) bookResponse {
	return bookResponse{
		ID:          b.ID.String(),
		AuthorID:    b.AuthorID.String(),
		GenreID:     b.GenreID.String(),
		Title:       b.Title,
		ISBN:        b.ISBN,
		CoverURL:    b.CoverURL,
		Description: b.Description,
		Language:    b.Language,
		Year:        b.Year,
		AvgRating:   b.AvgRating,
		TotalPages:  b.TotalPages,
	}
}
