package genre

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.GenreUseCase
}

func NewHandler(uc domainUseCase.GenreUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать жанр
// @Tags        genres
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createGenreRequest true "Данные жанра"
// @Success     201 {object} genreResponse
// @Failure     400 {object} map[string]string
// @Router      /admin/genres [post]
func (h *Handler) Create(c *gin.Context) {
	var req createGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	genre := &entity.Genre{
		Name: req.Name,
		Slug: req.Slug,
	}

	result, err := h.uc.Create(c.Request.Context(), genre)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toGenreResponse(result))
}

// @Summary     Получить жанр
// @Tags        genres
// @Produce     json
// @Param       id path string true "ID жанра"
// @Success     200 {object} genreResponse
// @Failure     404 {object} map[string]string
// @Router      /genres/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	genre, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toGenreResponse(genre))
}

// @Summary     Получить жанр по slug
// @Tags        genres
// @Produce     json
// @Param       slug path string true "Slug жанра"
// @Success     200 {object} genreResponse
// @Failure     404 {object} map[string]string
// @Router      /genres/slug/{slug} [get]
func (h *Handler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	genre, err := h.uc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toGenreResponse(genre))
}

// @Summary     Список жанров
// @Tags        genres
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Router      /genres [get]
func (h *Handler) List(c *gin.Context) {
	genres, err := h.uc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*genreResponse, len(genres))
	for i, g := range genres {
		resp := toGenreResponse(g)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Обновить жанр
// @Tags        genres
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string           true "ID жанра"
// @Param       input body updateGenreRequest true "Данные для обновления"
// @Success     200 {object} genreResponse
// @Failure     404 {object} map[string]string
// @Router      /admin/genres/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateGenreRequest
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
	if req.Slug != nil {
		existing.Slug = *req.Slug
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toGenreResponse(result))
}

// @Summary     Удалить жанр
// @Tags        genres
// @Security    BearerAuth
// @Param       id path string true "ID жанра"
// @Success     204
// @Router      /admin/genres/{id} [delete]
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

func toGenreResponse(g *entity.Genre) genreResponse {
	return genreResponse{
		ID:   g.ID.String(),
		Name: g.Name,
		Slug: g.Slug,
	}
}
