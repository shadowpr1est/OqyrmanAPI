package book_machine

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.BookMachineUseCase
}

func NewHandler(uc domainUseCase.BookMachineUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать книгомат
// @Tags        book-machines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createBookMachineRequest true "Данные книгомата"
// @Success     201 {object} bookMachineResponse
// @Router      /admin/book-machines [post]
func (h *Handler) Create(c *gin.Context) {
	var req createBookMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	machine := &entity.BookMachine{
		Name:    req.Name,
		Address: req.Address,
		Lat:     req.Lat,
		Lng:     req.Lng,
		Status:  entity.MachineStatus(req.Status),
	}

	result, err := h.uc.Create(c.Request.Context(), machine)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toBookMachineResponse(result))
}

// @Summary     Получить книгомат
// @Tags        book-machines
// @Produce     json
// @Param       id path string true "ID книгомата"
// @Success     200 {object} bookMachineResponse
// @Router      /book-machines/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	machine, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toBookMachineResponse(machine))
}

// @Summary     Список книгоматов
// @Tags        book-machines
// @Produce     json
// @Param       limit  query int false "Лимит"  default(20)
// @Param       offset query int false "Отступ" default(0)
// @Success     200 {object} listBookMachineResponse
// @Router      /book-machines [get]
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	machines, total, err := h.uc.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookMachineResponse, len(machines))
	for i, m := range machines {
		resp := toBookMachineResponse(m)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listBookMachineResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Книгоматы рядом
// @Tags        book-machines
// @Produce     json
// @Param       lat    query number true  "Широта"
// @Param       lng    query number true  "Долгота"
// @Param       radius query number false "Радиус км" default(5)
// @Success     200 {object} map[string]interface{}
// @Router      /book-machines/nearby [get]
func (h *Handler) ListNearby(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat"})
		return
	}

	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng"})
		return
	}

	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "5"), 64)

	machines, err := h.uc.ListNearby(c.Request.Context(), lat, lng, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*bookMachineResponse, len(machines))
	for i, m := range machines {
		resp := toBookMachineResponse(m)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Обновить книгомат
// @Tags        book-machines
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string                  true "ID книгомата"
// @Param       input body updateBookMachineRequest true "Данные для обновления"
// @Success     200 {object} bookMachineResponse
// @Router      /admin/book-machines/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateBookMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Address != nil {
		existing.Address = *req.Address
	}
	if req.Lat != nil {
		existing.Lat = *req.Lat
	}
	if req.Lng != nil {
		existing.Lng = *req.Lng
	}
	if req.Status != nil {
		existing.Status = entity.MachineStatus(*req.Status)
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toBookMachineResponse(result))
}

// @Summary     Удалить книгомат
// @Tags        book-machines
// @Security    BearerAuth
// @Param       id path string true "ID книгомата"
// @Success     204
// @Router      /admin/book-machines/{id} [delete]
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

func toBookMachineResponse(m *entity.BookMachine) bookMachineResponse {
	return bookMachineResponse{
		ID:      m.ID.String(),
		Name:    m.Name,
		Address: m.Address,
		Lat:     m.Lat,
		Lng:     m.Lng,
		Status:  string(m.Status),
	}
}
