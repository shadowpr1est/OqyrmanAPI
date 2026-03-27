package stats

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
	stats, err := h.uc.GetUserStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toUserStatsResponse(stats))
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
