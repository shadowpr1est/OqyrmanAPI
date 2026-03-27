package review

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.ReviewUseCase
}

func NewHandler(uc domainUseCase.ReviewUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать отзыв
// @Tags        reviews
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createReviewRequest true "Данные отзыва"
// @Success     201 {object} reviewResponse
// @Router      /reviews [post]
func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	var req createReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	review := &entity.Review{
		UserID: userID,
		BookID: bookID,
		Rating: req.Rating,
		Body:   req.Body,
	}

	result, err := h.uc.Create(c.Request.Context(), review)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toReviewResponse(result))
}

// @Summary     Получить отзыв
// @Tags        reviews
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID отзыва"
// @Success     200 {object} reviewResponse
// @Router      /reviews/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	review, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toReviewResponse(review))
}

// @Summary     Отзывы на книгу
// @Tags        reviews
// @Produce     json
// @Param       book_id path string true  "ID книги"
// @Param       limit   query int    false "Лимит"    default(20)
// @Param       offset  query int    false "Смещение" default(0)
// @Success     200 {object} listReviewResponse
// @Router      /reviews/book/{book_id} [get]
func (h *Handler) ListByBook(c *gin.Context) {
	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	limit := 20
	offset := 0
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}

	reviews, total, err := h.uc.ListByBook(c.Request.Context(), bookID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*reviewResponse, len(reviews))
	for i, r := range reviews {
		resp := toReviewResponse(r)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listReviewResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Мои отзывы
// @Tags        reviews
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} map[string]interface{}
// @Router      /reviews/user [get]
func (h *Handler) ListByUser(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	reviews, err := h.uc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]*reviewResponse, len(reviews))
	for i, r := range reviews {
		resp := toReviewResponse(r)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Обновить отзыв
// @Tags        reviews
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string            true "ID отзыва"
// @Param       input body updateReviewRequest true "Данные для обновления"
// @Success     200 {object} reviewResponse
// @Router      /reviews/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if req.Rating != nil {
		existing.Rating = *req.Rating
	}
	if req.Body != nil {
		existing.Body = *req.Body
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toReviewResponse(result))
}

// @Summary     Удалить отзыв
// @Tags        reviews
// @Security    BearerAuth
// @Param       id path string true "ID отзыва"
// @Success     204
// @Router      /reviews/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func toReviewResponse(r *entity.Review) reviewResponse {
	return reviewResponse{
		ID:        r.ID.String(),
		UserID:    r.UserID.String(),
		BookID:    r.BookID.String(),
		Rating:    r.Rating,
		Body:      r.Body,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
