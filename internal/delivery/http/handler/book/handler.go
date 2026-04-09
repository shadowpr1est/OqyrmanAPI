package book

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
)

type Handler struct {
	uc domainUseCase.BookUseCase
}

func NewHandler(uc domainUseCase.BookUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Создать книгу
// @Tags        books
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       author_id   formData string true  "ID автора"
// @Param       genre_id    formData string true  "ID жанра"
// @Param       title       formData string true  "Название"
// @Param       isbn        formData string false "ISBN"
// @Param       description formData string false "Описание"
// @Param       language    formData string false "Язык"
// @Param       year        formData int    false "Год"
// @Param       cover       formData file   false "Обложка (jpg, png)"
// @Success     201 {object} bookResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/books [post]
func (h *Handler) Create(c *gin.Context) {
	var req createBookRequest
	if err := c.ShouldBindWith(&req, binding.FormMultipart); err != nil {
		common.ValidationErr(c, err)
		return
	}

	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author_id"})
		return
	}
	genreID, err := uuid.Parse(req.GenreID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre_id"})
		return
	}

	book := &entity.Book{
		AuthorID:    authorID,
		GenreID:     genreID,
		Title:       req.Title,
		ISBN:        req.ISBN,
		Description: req.Description,
		Language:    req.Language,
		Year:        req.Year,
		TotalPages:  req.TotalPages,
	}

	var cover *fileupload.File
	if fh, err := c.FormFile("cover"); err == nil {
		f, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open cover file"})
			return
		}
		defer f.Close()
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		ct := http.DetectContentType(buf[:n])
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot process cover file"})
			return
		}
		switch ct {
		case "image/jpeg", "image/png", "image/webp":
			// ok
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "only jpeg, png, webp images are allowed"})
			return
		}
		cover = &fileupload.File{
			Filename:    fh.Filename,
			Reader:      f,
			Size:        fh.Size,
			ContentType: ct,
		}
	}

	result, err := h.uc.Create(c.Request.Context(), book, cover)
	if err != nil {
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
	c.JSON(http.StatusCreated, toBookViewResponse(view))
}

// @Summary     Загрузить обложку книги
// @Tags        books
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       id    path     string true "ID книги"
// @Param       cover formData file   true "Обложка (jpg, png)"
// @Success     200 {object} bookResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /admin/books/{id}/cover [post]
func (h *Handler) UploadCover(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	fh, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cover file is required"})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open cover file"})
		return
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	ct := http.DetectContentType(buf[:n])
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot process cover file"})
		return
	}
	switch ct {
	case "image/jpeg", "image/png", "image/webp":
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "only jpeg, png, webp images are allowed"})
		return
	}

	book, err := h.uc.UploadCover(c.Request.Context(), id, &fileupload.File{
		Filename:    fh.Filename,
		Reader:      f,
		Size:        fh.Size,
		ContentType: ct,
	})
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	view, err := h.uc.GetByIDView(c.Request.Context(), book.ID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, toBookViewResponse(view))
}

// @Summary     Получить книгу
// @Tags        books
// @Produce     json
// @Param       id path string true "ID книги"
// @Success     200 {object} bookViewResponse
// @Failure     404 {object} map[string]string
// @Router      /books/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	book, err := h.uc.GetByIDView(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "book not found")
		return
	}

	c.JSON(http.StatusOK, toBookViewResponse(book))
}

// @Summary     Список книг
// @Tags        books
// @Produce     json
// @Param       limit  query int false "Лимит"  default(20)
// @Param       offset query int false "Отступ" default(0)
// @Success     200 {object} listBookViewResponse
// @Router      /books [get]
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	if offset > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must not exceed 10000"})
		return
	}

	books, total, err := h.uc.ListView(c.Request.Context(), limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*bookViewResponse, len(books))
	for i, b := range books {
		resp := toBookViewResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listBookViewResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Книги по автору
// @Tags        books
// @Produce     json
// @Param       author_id path string true "ID автора"
// @Success     200 {object} map[string]interface{}
// @Router      /books/author/{author_id} [get]
func (h *Handler) ListByAuthor(c *gin.Context) {
	authorID, err := uuid.Parse(c.Param("author_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author_id"})
		return
	}

	books, err := h.uc.ListByAuthorView(c.Request.Context(), authorID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*bookViewResponse, len(books))
	for i, b := range books {
		resp := toBookViewResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Книги по жанру
// @Tags        books
// @Produce     json
// @Param       genre_id path string true "ID жанра"
// @Success     200 {object} map[string]interface{}
// @Router      /books/genre/{genre_id} [get]
func (h *Handler) ListByGenre(c *gin.Context) {
	genreID, err := uuid.Parse(c.Param("genre_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre_id"})
		return
	}

	books, err := h.uc.ListByGenreView(c.Request.Context(), genreID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*bookViewResponse, len(books))
	for i, b := range books {
		resp := toBookViewResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// @Summary     Поиск книг
// @Tags        books
// @Produce     json
// @Param       q      query string true  "Поисковый запрос"
// @Param       limit  query int    false "Лимит"    default(20)
// @Param       offset query int    false "Смещение" default(0)
// @Success     200 {object} listBookViewResponse
// @Failure     400 {object} map[string]string
// @Router      /books/search [get]
func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query param q is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	if offset > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must not exceed 10000"})
		return
	}
	if len(query) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query must not exceed 255 characters"})
		return
	}

	books, total, err := h.uc.SearchView(c.Request.Context(), query, limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*bookViewResponse, len(books))
	for i, b := range books {
		resp := toBookViewResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listBookViewResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Обновить книгу
// @Tags        books
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string          true "ID книги"
// @Param       input body updateBookRequest true "Данные для обновления"
// @Success     200 {object} bookResponse
// @Failure     404 {object} map[string]string
// @Router      /admin/books/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "book not found")
		return
	}

	if req.AuthorID != nil {
		authorID, err := uuid.Parse(*req.AuthorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid author_id"})
			return
		}
		existing.AuthorID = authorID
	}
	if req.GenreID != nil {
		genreID, err := uuid.Parse(*req.GenreID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre_id"})
			return
		}
		existing.GenreID = genreID
	}
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.ISBN != nil {
		existing.ISBN = *req.ISBN
	}
	if req.CoverURL != nil {
		existing.CoverURL = *req.CoverURL
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Language != nil {
		existing.Language = *req.Language
	}
	if req.Year != nil {
		existing.Year = *req.Year
	}
	if req.TotalPages != nil {
		existing.TotalPages = req.TotalPages
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
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
	c.JSON(http.StatusOK, toBookViewResponse(view))
}

// @Summary     Удалить книгу
// @Tags        books
// @Security    BearerAuth
// @Param       id path string true "ID книги"
// @Success     204
// @Router      /admin/books/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Популярные книги
// @Tags        books
// @Produce     json
// @Param       limit  query int false "Лимит"  default(20)
// @Param       offset query int false "Отступ" default(0)
// @Success     200 {object} listBookViewResponse
// @Router      /books/popular [get]
func (h *Handler) ListPopular(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	books, total, err := h.uc.ListPopularView(c.Request.Context(), limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*bookViewResponse, len(books))
	for i, b := range books {
		resp := toBookViewResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, listBookViewResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Похожие книги
// @Tags        books
// @Produce     json
// @Param       id    path string true  "ID книги"
// @Param       limit query int   false "Лимит" default(10)
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Router      /books/{id}/similar [get]
func (h *Handler) ListSimilar(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	books, err := h.uc.ListSimilarView(c.Request.Context(), id, limit)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	items := make([]*bookViewResponse, len(books))
	for i, b := range books {
		resp := toBookViewResponse(b)
		items[i] = &resp
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func toBookResponse(b *entity.Book) bookResponse {
	return bookResponse{
		ID:          b.ID.String(),
		AuthorID:    b.AuthorID.String(),
		GenreID:     b.GenreID.String(),
		Title:       b.Title,
		ISBN:        b.ISBN,
		CoverURL:    b.CoverURL,
		Description: b.Description,
		Language:    b.Language,
		Year:        b.Year,
		AvgRating:   b.AvgRating,
		TotalPages:  b.TotalPages,
	}
}

func toBookViewResponse(v *entity.BookView) bookViewResponse {
	return bookViewResponse{
		ID:          v.ID.String(),
		Title:       v.Title,
		ISBN:        v.ISBN,
		CoverURL:    v.CoverURL,
		Description: v.Description,
		Language:    v.Language,
		Year:        v.Year,
		AvgRating:   v.AvgRating,
		TotalPages:  v.TotalPages,
		BookFile: func() *common.BookFileRef {
			if v.BookFileID == nil {
				return nil
			}
			return &common.BookFileRef{
				ID:      *v.BookFileID,
				BookID:  *v.BookFileBookID,
				Format:  *v.BookFileFormat,
				FileURL: *v.BookFileUrl,
			}
		}(),
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
	}
}
