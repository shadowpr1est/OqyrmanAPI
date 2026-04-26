package library

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
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
		common.ValidationErr(c, err)
		return
	}

	library := &entity.Library{
		Name:     req.Name,
		Address:  req.Address,
		Lat:      req.Lat,
		Lng:      req.Lng,
		Phone:    req.Phone,
		PhotoURL: req.PhotoURL,
	}

	result, err := h.uc.Create(c.Request.Context(), library)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
		common.NotFound(c, "library not found")
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
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
	if lat < -90 || lat > 90 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat must be between -90 and 90"})
		return
	}

	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng"})
		return
	}
	if lng < -180 || lng > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lng must be between -180 and 180"})
		return
	}

	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "5"), 64)
	if radius <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "radius must be positive"})
		return
	}

	libraries, err := h.uc.ListNearby(c.Request.Context(), lat, lng, radius)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
		common.ValidationErr(c, err)
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		common.NotFound(c, "library not found")
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
	if req.PhotoURL != nil {
		existing.PhotoURL = *req.PhotoURL
	}

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "library not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Загрузить фото библиотеки
// @Tags        libraries
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       id    path     string true "ID библиотеки"
// @Param       photo formData file   true "Фото (jpg, png, webp)"
// @Success     200 {object} libraryResponse
// @Router      /admin/libraries/{id}/photo [post]
func (h *Handler) UploadPhoto(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	fh, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photo file is required"})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open photo file"})
		return
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	ct := http.DetectContentType(buf[:n])
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot process photo file"})
		return
	}
	switch ct {
	case "image/jpeg", "image/png", "image/webp":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "only jpeg, png, webp images are allowed"})
		return
	}

	library, err := h.uc.UploadPhoto(c.Request.Context(), id, &fileupload.File{
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
	c.JSON(http.StatusOK, toLibraryResponse(library))
}

func toLibraryResponse(l *entity.Library) libraryResponse {
	return libraryResponse{
		ID:       l.ID.String(),
		Name:     l.Name,
		Address:  l.Address,
		Lat:      l.Lat,
		Lng:      l.Lng,
		Phone:    l.Phone,
		PhotoURL: l.PhotoURL,
	}
}
