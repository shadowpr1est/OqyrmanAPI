package stats

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
