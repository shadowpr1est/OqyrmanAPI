package event

import "github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"

type createEventRequest struct {
	Title       string  `form:"title"       binding:"required"`
	Description *string `form:"description"`
	Location    *string `form:"location"`
	StartsAt    string  `form:"starts_at"   binding:"required"`
	EndsAt      string  `form:"ends_at"     binding:"required"`
}

type updateEventRequest struct {
	Title       string  `form:"title"       binding:"required"`
	Description *string `form:"description"`
	Location    *string `form:"location"`
	StartsAt    string  `form:"starts_at"   binding:"required"`
	EndsAt      string  `form:"ends_at"     binding:"required"`
}

type eventResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	CoverURL    *string `json:"cover_url,omitempty"`
	Location    *string `json:"location,omitempty"`
	StartsAt    string  `json:"starts_at"`
	EndsAt      string  `json:"ends_at"`
	CreatedAt   string  `json:"created_at"`
}

func toEventResponse(e *entity.Event) eventResponse {
	return eventResponse{
		ID:          e.ID.String(),
		Title:       e.Title,
		Description: e.Description,
		CoverURL:    e.CoverURL,
		Location:    e.Location,
		StartsAt:    e.StartsAt.Format("2006-01-02T15:04:05Z"),
		EndsAt:      e.EndsAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
