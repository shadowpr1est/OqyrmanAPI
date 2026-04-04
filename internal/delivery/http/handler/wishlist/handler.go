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
// @Param       input body addWishlistRequest true "ID книги"
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

	result, err := h.uc.Add(c.Request.Context(), userID, bookID)
	if err != nil {
		if errors.Is(err, entity.ErrDuplicateWishlist) {
			c.JSON(http.StatusConflict, gin.H{"error": "book already in wishlist"})
			return
		}
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
// @Success     200 {object} map[string]interface{}
// @Router      /wishlist [get]
func (h *Handler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)

	items, err := h.uc.ListByUserView(c.Request.Context(), userID)
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
// @Success     200 {object} map[string]bool
// @Router      /wishlist/{book_id}/exists [get]
func (h *Handler) Exists(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bookID, err := uuid.Parse(c.Param("book_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book_id"})
		return
	}

	exists, err := h.uc.ExistsByUserAndBook(c.Request.Context(), userID, bookID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func toWishlistResponse(w *entity.Wishlist) wishlistResponse {
	return wishlistResponse{
		ID:      w.ID.String(),
		BookID:  w.BookID.String(),
		AddedAt: w.AddedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toWishlistViewResponse(v *entity.WishlistView) wishlistViewResponse {
	return wishlistViewResponse{
		ID:      v.ID.String(),
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
