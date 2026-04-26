package author

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.AuthorUseCase
}

func NewHandler(uc domainUseCase.AuthorUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать автора
// @Tags        authors
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createAuthorRequest true "Данные автора"
// @Success     201 {object} authorResponse
// @Failure     400 {object} map[string]string
// @Router      /admin/authors [post]
func (h *Handler) Create(c *gin.Context) {
	var req createAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	author := &entity.Author{
		Name:     req.Name,
		Bio:      req.Bio,
		BioKK:    req.BioKK,
		PhotoURL: req.PhotoURL,
	}

	if req.BirthDate != nil {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birth_date format, use YYYY-MM-DD"})
			return
		}
		author.BirthDate = &t
	}

	if req.DeathDate != nil {
		t, err := time.Parse("2006-01-02", *req.DeathDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid death_date format, use YYYY-MM-DD"})
			return
		}
		author.DeathDate = &t
	}

	result, err := h.uc.Create(c.Request.Context(), author)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toAuthorResponse(result))
}

// @Summary     Получить автора
// @Tags        authors
// @Produce     json
// @Param       id path string true "ID автора"
// @Success     200 {object} authorResponse
// @Failure     404 {object} map[string]string
// @Router      /authors/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	author, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "author not found")
		return
	}

	c.JSON(http.StatusOK, toAuthorResponse(author))
}

// @Summary     Список авторов
// @Tags        authors
// @Produce     json
// @Param       limit  query int false "Лимит"  default(20)
// @Param       offset query int false "Отступ" default(0)
// @Success     200 {object} listAuthorResponse
// @Failure     500 {object} map[string]string
// @Router      /authors [get]
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	authors, total, err := h.uc.List(c.Request.Context(), limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*authorResponse, len(authors))
	for i, a := range authors {
		resp := toAuthorResponse(a)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listAuthorResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Обновить автора
// @Tags        authors
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string            true "ID автора"
// @Param       input body updateAuthorRequest true "Данные для обновления"
// @Success     200 {object} authorResponse
// @Failure     404 {object} map[string]string
// @Router      /admin/authors/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "author not found")
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Bio != nil {
		existing.Bio = *req.Bio
	}
	if req.BioKK != nil {
		existing.BioKK = *req.BioKK
	}
	if req.PhotoURL != nil {
		existing.PhotoURL = *req.PhotoURL
	}
	if req.BirthDate != nil {
		t, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid birth_date format, use YYYY-MM-DD"})
			return
		}
		existing.BirthDate = &t
	}
	if req.DeathDate != nil {
		t, err := time.Parse("2006-01-02", *req.DeathDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid death_date format, use YYYY-MM-DD"})
			return
		}
		existing.DeathDate = &t
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, toAuthorResponse(result))
}

// @Summary     Удалить автора
// @Tags        authors
// @Security    BearerAuth
// @Param       id path string true "ID автора"
// @Success     204
// @Failure     500 {object} map[string]string
// @Router      /admin/authors/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "author not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Поиск авторов
// @Tags        authors
// @Produce     json
// @Param       q      query string true  "Поисковый запрос"
// @Param       limit  query int    false "Лимит"    default(20)
// @Param       offset query int    false "Смещение" default(0)
// @Success     200 {object} listAuthorResponse
// @Failure     400 {object} map[string]string
// @Router      /authors/search [get]
func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	authors, total, err := h.uc.Search(c.Request.Context(), q, limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*authorResponse, len(authors))
	for i, a := range authors {
		resp := toAuthorResponse(a)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listAuthorResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func toAuthorResponse(a *entity.Author) authorResponse {
	resp := authorResponse{
		ID:       a.ID.String(),
		Name:     a.Name,
		Bio:      a.Bio,
		BioKK:    a.BioKK,
		PhotoURL: a.PhotoURL,
	}
	if a.BirthDate != nil {
		s := a.BirthDate.Format("2006-01-02")
		resp.BirthDate = &s
	}
	if a.DeathDate != nil {
		s := a.DeathDate.Format("2006-01-02")
		resp.DeathDate = &s
	}
	return resp
}
