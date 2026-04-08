package wishlist

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.WishlistUseCase
}

func NewHandler(uc domainUseCase.WishlistUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Добавить в вишлист
// @Tags        wishlist
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body addWishlistRequest true "ID книги и статус"
// @Success     201 {object} wishlistResponse
// @Router      /wishlist [post]
func (h *Handler) Add(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req addWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	bookID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	status := entity.ShelfWantToRead
	if req.Status != "" {
		s := entity.ShelfStatus(req.Status)
		switch s {
		case entity.ShelfWantToRead, entity.ShelfReading, entity.ShelfFinished:
			status = s
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
	}

	result, err := h.uc.Add(c.Request.Context(), userID, bookID, status)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toWishlistResponse(result))
}

// @Summary     Удалить из вишлиста
// @Tags        wishlist
// @Security    BearerAuth
// @Param       book_id path string true "ID книги"
// @Success     204
// @Router      /wishlist/{book_id} [delete]
func (h *Handler) Remove(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	if err := h.uc.Remove(c.Request.Context(), userID, bookID); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Мой вишлист
// @Tags        wishlist
// @Security    BearerAuth
// @Produce     json
// @Param       status query string false "Фильтр по статусу (want_to_read, reading, finished)"
// @Success     200 {object} map[string]interface{}
// @Router      /wishlist [get]
func (h *Handler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var statusFilter *entity.ShelfStatus
	if s := c.Query("status"); s != "" {
		st := entity.ShelfStatus(s)
		switch st {
		case entity.ShelfWantToRead, entity.ShelfReading, entity.ShelfFinished:
			statusFilter = &st
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status filter"})
			return
		}
	}

	items, err := h.uc.ListByUserView(c.Request.Context(), userID, statusFilter)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	resp := make([]*wishlistViewResponse, len(items))
	for i, w := range items {
		r := toWishlistViewResponse(w)
		resp[i] = &r
	}

	c.JSON(http.StatusOK, gin.H{"items": resp})
}

// @Summary     Проверить наличие в вишлисте
// @Tags        wishlist
// @Security    BearerAuth
// @Produce     json
// @Param       book_id path string true "ID книги"
// @Success     200 {object} map[string]interface{}
// @Router      /wishlist/{book_id}/exists [get]
func (h *Handler) Exists(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	status, err := h.uc.GetStatusByUserAndBook(c.Request.Context(), userID, bookID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	if status == nil {
		c.JSON(http.StatusOK, gin.H{"exists": false, "status": nil})
	} else {
		c.JSON(http.StatusOK, gin.H{"exists": true, "status": string(*status)})
	}
}

// @Summary     Обновить статус на полке
// @Tags        wishlist
// @Security    BearerAuth
// @Accept      json
// @Param       book_id path string true "ID книги"
// @Param       input body updateStatusRequest true "Новый статус"
// @Success     204
// @Router      /wishlist/{book_id}/status [patch]
func (h *Handler) UpdateStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	s := entity.ShelfStatus(req.Status)
	switch s {
	case entity.ShelfWantToRead, entity.ShelfReading, entity.ShelfFinished:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	if err := h.uc.UpdateStatus(c.Request.Context(), userID, bookID, s); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not on shelf"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func toWishlistResponse(w *entity.Wishlist) wishlistResponse {
	return wishlistResponse{
		ID:      w.ID.String(),
		BookID:  w.BookID.String(),
		Status:  string(w.Status),
		AddedAt: w.AddedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toWishlistViewResponse(v *entity.WishlistView) wishlistViewResponse {
	return wishlistViewResponse{
		ID:      v.ID.String(),
		Status:  string(v.Status),
		AddedAt: v.AddedAt.Format("2006-01-02T15:04:05Z"),
		Book: common.BookRef{
			ID:        v.BookID.String(),
			Title:     v.BookTitle,
			CoverURL:  v.BookCoverURL,
			AvgRating: v.BookAvgRating,
			Author: common.AuthorRef{
				ID:   v.AuthorID.String(),
				Name: v.AuthorName,
			},
			Genre: common.GenreRef{
				ID:   v.GenreID.String(),
				Name: v.GenreName,
			},
		},
	}
}
