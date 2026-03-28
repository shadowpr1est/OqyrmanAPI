package stats

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.StatsUseCase
}

func NewHandler(uc domainUseCase.StatsUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Статистика платформы
// @Tags        stats
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} statsResponse
// @Router      /admin/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.uc.GetStats(c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toStatsResponse(stats))
}

// @Summary     Моя статистика
// @Tags        stats
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} userStatsResponse
// @Router      /users/me/stats [get]
func (h *Handler) GetUserStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	stats, err := h.uc.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toUserStatsResponse(stats))
}

// @Summary     Статистика библиотеки (staff)
// @Tags        staff
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} libraryStatsResponse
// @Failure     403 {object} map[string]string
// @Router      /staff/library/stats [get]
func (h *Handler) GetLibraryStats(c *gin.Context) {
	libraryID := middleware.GetLibraryID(c)
	if libraryID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "no library assigned"})
		return
	}
	stats, err := h.uc.GetLibraryStats(c.Request.Context(), *libraryID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toLibraryStatsResponse(stats))
}

func toStatsResponse(s *entity.Stats) statsResponse {
	return statsResponse{
		UsersTotal:          s.UsersTotal,
		BooksTotal:          s.BooksTotal,
		AuthorsTotal:        s.AuthorsTotal,
		ReservationsActive:  s.ReservationsActive,
		ReservationsPending: s.ReservationsPending,
		ReservationsTotal:   s.ReservationsTotal,
		ReviewsTotal:        s.ReviewsTotal,
	}
}

func toUserStatsResponse(s *entity.UserStats) userStatsResponse {
	return userStatsResponse{
		BooksRead:          s.BooksRead,
		ActiveReservations: s.ActiveReservations,
		ReviewsGiven:       s.ReviewsGiven,
		WishlistCount:      s.WishlistCount,
	}
}

func toLibraryStatsResponse(s *entity.LibraryStats) libraryStatsResponse {
	return libraryStatsResponse{
		TotalBooks:            s.TotalBooks,
		AvailableBooks:        s.AvailableBooks,
		TotalReservations:     s.TotalReservations,
		ActiveReservations:    s.ActiveReservations,
		PendingReservations:   s.PendingReservations,
		CompletedReservations: s.CompletedReservations,
		CancelledReservations: s.CancelledReservations,
	}
}
