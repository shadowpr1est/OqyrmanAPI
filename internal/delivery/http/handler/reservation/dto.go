package reservation

import "strconv"
import "errors"
import "github.com/gin-gonic/gin"
import "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"

type createReservationRequest struct {
	LibraryBookID string `json:"library_book_id" binding:"required"`
	DueDate       string `json:"due_date" binding:"required"`
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type extendReservationRequest struct {
	DueDate string `json:"due_date" binding:"required"`
}

type reservationViewResponse struct {
	ID            string                    `json:"id"`
	Status        string                    `json:"status"`
	ReservedAt    string                    `json:"reserved_at"`
	DueDate       string                    `json:"due_date"`
	ReturnedAt    *string                   `json:"returned_at,omitempty"`
	LibraryBookID string                    `json:"library_book_id"`
	Book          common.ReservationBookRef `json:"book"`
	Library       common.LibraryRef         `json:"library"`
	// User присутствует только в staff/admin контексте — при просмотре чужих броней.
	User *common.UserRef `json:"user,omitempty"`
}

func parsePagination(c *gin.Context) (limit, offset int, err error) {
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 {
		return 0, 0, errors.New("limit must be greater than 0")
	}
	if limit > 100 {
		return 0, 0, errors.New("limit must not exceed 100")
	}
	if offset < 0 {
		return 0, 0, errors.New("offset must be >= 0")
	}
	if offset > 10000 {
		return 0, 0, errors.New("offset must not exceed 10000")
	}
	return limit, offset, nil
}
