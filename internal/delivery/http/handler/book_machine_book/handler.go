package book_machine_book

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.BookMachineBookUseCase
}

func NewHandler(uc domainUseCase.BookMachineBookUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Добавить книгу в книгомат
// @Tags        book-machine-books
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createBookMachineBookRequest true "Данные"
// @Success     201 {object} bookMachineBookResponse
// @Router      /admin/book-machine-books [post]
func (h *Handler) Create(c *gin.Context) {
	var req createBookMachineBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	machineID, err := uuid.Parse(req.MachineID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid machine_id"})
		return
	}

	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	mb := &entity.BookMachineBook{
		MachineID:       machineID,
		BookID:          bookID,
		TotalCopies:     req.TotalCopies,
		AvailableCopies: req.AvailableCopies,
	}

	result, err := h.uc.Create(c.Request.Context(), mb)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toBookMachineBookResponse(result))
}

// @Summary     Получить запись
// @Tags        book-machine-books
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID"
// @Success     200 {object} bookMachineBookResponse
// @Router      /book-machine-books/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	mb, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toBookMachineBookResponse(mb))
}

// @Summary     Книги в книгомате
// @Tags        book-machine-books
// @Security    BearerAuth
// @Produce     json
// @Param       machine_id path string true "ID книгомата"
// @Success     200 {object} map[string]interface{}
// @Router      /book-machine-books/machine/{machine_id} [get]
func (h *Handler) ListByMachine(c *gin.Context) {
	machineID, err := uuid.Parse(c.Param("machine_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid machine_id"})
		return
	}

	items, err := h.uc.ListByMachine(c.Request.Context(), machineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]*bookMachineBookResponse, len(items))
	for i, mb := range items {
		r := toBookMachineBookResponse(mb)
		resp[i] = &r
	}

	c.JSON(http.StatusOK, gin.H{"items": resp})
}

// @Summary     Книгоматы с книгой
// @Tags        book-machine-books
// @Security    BearerAuth
// @Produce     json
// @Param       book_id path string true "ID книги"
// @Success     200 {object} map[string]interface{}
// @Router      /book-machine-books/book/{book_id} [get]
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

	resp := make([]*bookMachineBookResponse, len(items))
	for i, mb := range items {
		r := toBookMachineBookResponse(mb)
		resp[i] = &r
	}

	c.JSON(http.StatusOK, gin.H{"items": resp})
}

// @Summary     Обновить копии
// @Tags        book-machine-books
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string                     true "ID"
// @Param       input body updateBookMachineBookRequest true "Данные"
// @Success     200 {object} bookMachineBookResponse
// @Router      /admin/book-machine-books/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateBookMachineBookRequest
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

	c.JSON(http.StatusOK, toBookMachineBookResponse(result))
}

// @Summary     Удалить запись
// @Tags        book-machine-books
// @Security    BearerAuth
// @Param       id path string true "ID"
// @Success     204
// @Router      /admin/book-machine-books/{id} [delete]
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

func toBookMachineBookResponse(mb *entity.BookMachineBook) bookMachineBookResponse {
	return bookMachineBookResponse{
		ID:              mb.ID.String(),
		MachineID:       mb.MachineID.String(),
		BookID:          mb.BookID.String(),
		TotalCopies:     mb.TotalCopies,
		AvailableCopies: mb.AvailableCopies,
	}
}
