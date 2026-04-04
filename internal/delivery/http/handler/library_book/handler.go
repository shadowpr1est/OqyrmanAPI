package library_book

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
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
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, toLibraryBookResponse(result))
}

// @Summary     Получить запись
// @Tags        library-books
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID"
// @Success     200 {object} libraryBookViewResponse
// @Router      /library-books/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	v, err := h.uc.GetByIDView(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "library book not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toLibraryBookViewResponse(v))
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

	items, err := h.uc.ListByLibraryView(c.Request.Context(), libraryID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]*libraryBookViewResponse, len(items))
	for i, v := range items {
		r := toLibraryBookViewResponse(v)
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

	items, err := h.uc.ListByBookView(c.Request.Context(), bookID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]*libraryBookViewResponse, len(items))
	for i, v := range items {
		r := toLibraryBookViewResponse(v)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": common.ValidationError(err)})
		return
	}

	// UpdateCopies applies COALESCE in SQL — no prior read needed, avoids TOCTOU.
	result, err := h.uc.UpdateCopies(c.Request.Context(), id, req.TotalCopies, req.AvailableCopies)
	if err != nil {
		if errors.Is(err, entity.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "library book not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
		if errors.Is(err, entity.ErrActiveReservationsExist) {
			c.JSON(http.StatusConflict, gin.H{"error": "cannot delete: active or pending reservations exist"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Поиск книг в библиотеке (для staff)
// @Tags        library-books
// @Security    BearerAuth
// @Produce     json
// @Param       q         query string false "Поиск по названию/автору"
// @Param       genre_id  query string false "Фильтр по жанру (UUID)"
// @Param       available query bool   false "Только доступные"
// @Param       limit     query int    false "Лимит (default 20)"
// @Param       offset    query int    false "Смещение (default 0)"
// @Success     200 {object} libraryBookSearchResponse
// @Router      /staff/books/search [get]
func (h *Handler) SearchInLibrary(c *gin.Context) {
	libraryID := middleware.GetLibraryID(c)
	if libraryID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "library_id not set for this staff"})
		return
	}

	q := c.Query("q")

	var genreID *uuid.UUID
	if raw := c.Query("genre_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre_id"})
			return
		}
		genreID = &parsed
	}

	onlyAvailable := c.Query("available") == "true"

	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}

	offset := 0
	if raw := c.Query("offset"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v >= 0 {
			offset = v
		}
	}

	items, total, err := h.uc.SearchInLibrary(c.Request.Context(), *libraryID, q, genreID, onlyAvailable, limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	resp := make([]*libraryBookSearchItem, len(items))
	for i, item := range items {
		resp[i] = &libraryBookSearchItem{
			LibraryBookID:   item.LibraryBookID.String(),
			BookID:          item.BookID.String(),
			Title:           item.Title,
			Author:          item.Author,
			Genre:           item.Genre,
			CoverURL:        item.CoverURL,
			Year:            item.Year,
			TotalCopies:     item.TotalCopies,
			AvailableCopies: item.AvailableCopies,
			IsAvailable:     item.IsAvailable,
		}
	}

	c.JSON(http.StatusOK, libraryBookSearchResponse{
		Items:  resp,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
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

func toLibraryBookViewResponse(v *entity.LibraryBookView) libraryBookViewResponse {
	return libraryBookViewResponse{
		ID: v.ID.String(),
		Library: common.LibraryRef{
			ID:      v.LibraryID.String(),
			Name:    v.LibraryName,
			Address: v.LibraryAddress,
			Lat:     v.LibraryLat,
			Lng:     v.LibraryLng,
			Phone:   v.LibraryPhone,
		},
		Book: common.BookRef{
			ID:          v.BookID.String(),
			Title:       v.BookTitle,
			ISBN:        v.BookISBN,
			CoverURL:    v.BookCoverURL,
			Description: v.BookDescription,
			Language:    v.BookLanguage,
			Year:        v.BookYear,
			TotalPages:  v.BookTotalPages,
			AvgRating:   v.BookAvgRating,
			Author: common.AuthorRef{
				ID:        v.AuthorID.String(),
				Name:      v.AuthorName,
				Bio:       v.AuthorBio,
				BirthDate: v.AuthorBirthDate,
				DeathDate: v.AuthorDeathDate,
				PhotoURL:  v.AuthorPhotoURL,
			},
			Genre: common.GenreRef{
				ID:   v.GenreID.String(),
				Name: v.GenreName,
				Slug: v.GenreSlug,
			},
		},
		TotalCopies:     v.TotalCopies,
		AvailableCopies: v.AvailableCopies,
	}
}
