package review

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
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
	userID := middleware.GetUserID(c)

	var req createReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
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
		if errors.Is(err, entity.ErrForbidden) {
			common.Forbidden(c)
			return
		}
		if errors.Is(err, entity.ErrValidation) {
			common.BadRequest(c, common.CodeValidationError, "invalid input")
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	view, err := h.uc.GetByIDView(c.Request.Context(), result.ID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	c.JSON(http.StatusCreated, toReviewViewResponse(view))
}

// @Summary     Получить отзыв
// @Tags        reviews
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID отзыва"
// @Success     200 {object} reviewViewResponse
// @Router      /reviews/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	review, err := h.uc.GetByIDView(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "review not found")
		return
	}
	c.JSON(http.StatusOK, toReviewViewResponse(review))
}

// @Summary     Отзывы на книгу
// @Tags        reviews
// @Produce     json
// @Param       book_id path string true  "ID книги"
// @Param       limit   query int    false "Лимит"    default(20)
// @Param       offset  query int    false "Смещение" default(0)
// @Success     200 {object} listReviewViewResponse
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

	reviews, total, err := h.uc.ListByBookView(c.Request.Context(), bookID, limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*reviewViewResponse, len(reviews))
	for i, r := range reviews {
		resp := toReviewViewResponse(r)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listReviewViewResponse{
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
	userID := middleware.GetUserID(c)

	reviews, err := h.uc.ListByUserView(c.Request.Context(), userID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*reviewViewResponse, len(reviews))
	for i, r := range reviews {
		resp := toReviewViewResponse(r)
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
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "review not found")
		return
	}

	if req.Rating != nil {
		existing.Rating = *req.Rating
	}
	if req.Body != nil {
		existing.Body = *req.Body
	}

	result, err := h.uc.Update(c.Request.Context(), existing, userID)
	if err != nil {
		if errors.Is(err, entity.ErrForbidden) {
			common.Forbidden(c)
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	view, err := h.uc.GetByIDView(c.Request.Context(), result.ID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, toReviewViewResponse(view))
}

// @Summary     Удалить отзыв
// @Tags        reviews
// @Security    BearerAuth
// @Param       id path string true "ID отзыва"
// @Success     204
// @Router      /reviews/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
			return
		}
		if errors.Is(err, entity.ErrForbidden) {
			common.Forbidden(c)
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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

func toReviewViewResponse(v *entity.ReviewView) reviewViewResponse {
	return reviewViewResponse{
		ID:            v.ID.String(),
		BookID:        v.BookID.String(),
		BookTitle:     v.BookTitle,
		UserID:        v.UserID.String(),
		UserName:      v.UserName,
		UserSurname:   v.UserSurname,
		UserAvatarURL: v.UserAvatarURL,
		Rating:        v.Rating,
		Body:          v.Body,
		CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
