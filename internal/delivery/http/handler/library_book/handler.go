package library_book

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.LibraryBookUseCase
}

func NewHandler(uc domainUseCase.LibraryBookUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Добавить книгу в библиотеку
// @Tags        library-books
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createLibraryBookRequest true "Данные"
// @Success     201 {object} libraryBookResponse
// @Router      /admin/library-books [post]
func (h *Handler) Create(c *gin.Context) {
	var req createLibraryBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	libraryID, err := uuid.Parse(req.LibraryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_id"})
		return
	}

	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	lb := &entity.LibraryBook{
		LibraryID:       libraryID,
		BookID:          bookID,
		TotalCopies:     req.TotalCopies,
		AvailableCopies: req.AvailableCopies,
	}

	result, err := h.uc.Create(c.Request.Context(), lb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toLibraryBookResponse(result))
}

// @Summary     Получить запись
// @Tags        library-books
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID"
// @Success     200 {object} libraryBookResponse
// @Router      /library-books/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	lb, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toLibraryBookResponse(lb))
}

// @Summary     Книги в библиотеке
// @Tags        library-books
// @Security    BearerAuth
// @Produce     json
// @Param       library_id path string true "ID библиотеки"
// @Success     200 {object} map[string]interface{}
// @Router      /library-books/library/{library_id} [get]
func (h *Handler) ListByLibrary(c *gin.Context) {
	libraryID, err := uuid.Parse(c.Param("library_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_id"})
		return
	}

	items, err := h.uc.ListByLibrary(c.Request.Context(), libraryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]*libraryBookResponse, len(items))
	for i, lb := range items {
		r := toLibraryBookResponse(lb)
		resp[i] = &r
	}

	c.JSON(http.StatusOK, gin.H{"items": resp})
}

// @Summary     Библиотеки с книгой
// @Tags        library-books
// @Security    BearerAuth
// @Produce     json
// @Param       book_id path string true "ID книги"
// @Success     200 {object} map[string]interface{}
// @Router      /library-books/book/{book_id} [get]
func (h *Handler) ListByBook(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	items, err := h.uc.ListByBook(c.Request.Context(), bookID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]*libraryBookResponse, len(items))
	for i, lb := range items {
		r := toLibraryBookResponse(lb)
		resp[i] = &r
	}

	c.JSON(http.StatusOK, gin.H{"items": resp})
}

// @Summary     Обновить копии
// @Tags        library-books
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string                 true "ID"
// @Param       input body updateLibraryBookRequest true "Данные"
// @Success     200 {object} libraryBookResponse
// @Router      /admin/library-books/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateLibraryBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.TotalCopies != nil {
		existing.TotalCopies = *req.TotalCopies
	}
	if req.AvailableCopies != nil {
		existing.AvailableCopies = *req.AvailableCopies
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toLibraryBookResponse(result))
}

// @Summary     Удалить запись
// @Tags        library-books
// @Security    BearerAuth
// @Param       id path string true "ID"
// @Success     204
// @Router      /admin/library-books/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "library book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func toLibraryBookResponse(lb *entity.LibraryBook) libraryBookResponse {
	return libraryBookResponse{
		ID:              lb.ID.String(),
		LibraryID:       lb.LibraryID.String(),
		BookID:          lb.BookID.String(),
		TotalCopies:     lb.TotalCopies,
		AvailableCopies: lb.AvailableCopies,
	}
}
