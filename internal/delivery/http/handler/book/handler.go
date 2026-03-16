package book

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.BookUseCase
}

func NewHandler(uc domainUseCase.BookUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать книгу
// @Tags        books
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createBookRequest true "Данные книги"
// @Success     201 {object} bookResponse
// @Failure     400 {object} map[string]string
// @Router      /admin/books [post]
func (h *Handler) Create(c *gin.Context) {
	var req createBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
		CoverURL:    req.CoverURL,
		Description: req.Description,
		Language:    req.Language,
		Year:        req.Year,
		AvgRating:   req.AvgRating,
	}

	result, err := h.uc.Create(c.Request.Context(), book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toBookResponse(result))
}

// @Summary     Получить книгу
// @Tags        books
// @Security    BearerAuth
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
// @Security    BearerAuth
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
// @Security    BearerAuth
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
// @Security    BearerAuth
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
// @Security    BearerAuth
// @Produce     json
// @Param       q      query string true  "Поисковый запрос"
// @Param       limit  query int    false "Лимит"  default(20)
// @Param       offset query int    false "Отступ" default(0)
// @Success     200 {object} listBookResponse
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
	if req.AvgRating != nil {
		existing.AvgRating = *req.AvgRating
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
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
	}
}
