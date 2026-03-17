package library

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.LibraryUseCase
}

func NewHandler(uc domainUseCase.LibraryUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать библиотеку
// @Tags        libraries
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createLibraryRequest true "Данные библиотеки"
// @Success     201 {object} libraryResponse
// @Router      /admin/libraries [post]
func (h *Handler) Create(c *gin.Context) {
	var req createLibraryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	library := &entity.Library{
		Name:    req.Name,
		Address: req.Address,
		Lat:     req.Lat,
		Lng:     req.Lng,
		Phone:   req.Phone,
	}

	result, err := h.uc.Create(c.Request.Context(), library)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toLibraryResponse(result))
}

// @Summary     Получить библиотеку
// @Tags        libraries
// @Produce     json
// @Param       id path string true "ID библиотеки"
// @Success     200 {object} libraryResponse
// @Router      /libraries/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	library, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toLibraryResponse(library))
}

// @Summary     Список библиотек
// @Tags        libraries
// @Produce     json
// @Param       limit  query int false "Лимит"  default(20)
// @Param       offset query int false "Отступ" default(0)
// @Success     200 {object} listLibraryResponse
// @Router      /libraries [get]
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	libraries, total, err := h.uc.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*libraryResponse, len(libraries))
	for i, l := range libraries {
		resp := toLibraryResponse(l)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listLibraryResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Библиотеки рядом
// @Tags        libraries
// @Produce     json
// @Param       lat    query number true  "Широта"
// @Param       lng    query number true  "Долгота"
// @Param       radius query number false "Радиус км" default(5)
// @Success     200 {object} map[string]interface{}
// @Router      /libraries/nearby [get]
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

	libraries, err := h.uc.ListNearby(c.Request.Context(), lat, lng, radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*libraryResponse, len(libraries))
	for i, l := range libraries {
		resp := toLibraryResponse(l)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Обновить библиотеку
// @Tags        libraries
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string             true "ID библиотеки"
// @Param       input body updateLibraryRequest true "Данные для обновления"
// @Success     200 {object} libraryResponse
// @Router      /admin/libraries/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateLibraryRequest
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
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toLibraryResponse(result))
}

// @Summary     Удалить библиотеку
// @Tags        libraries
// @Security    BearerAuth
// @Param       id path string true "ID библиотеки"
// @Success     204
// @Router      /admin/libraries/{id} [delete]
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

func toLibraryResponse(l *entity.Library) libraryResponse {
	return libraryResponse{
		ID:      l.ID.String(),
		Name:    l.Name,
		Address: l.Address,
		Lat:     l.Lat,
		Lng:     l.Lng,
		Phone:   l.Phone,
	}
}
