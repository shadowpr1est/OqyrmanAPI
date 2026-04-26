package event

import "github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"

type createEventRequest struct {
	Title         string  `form:"title"          binding:"required"`
	TitleKK       *string `form:"title_kk"`
	Description   *string `form:"description"`
	DescriptionKK *string `form:"description_kk"`
	Location      *string `form:"location"`
	StartsAt      string  `form:"starts_at"      binding:"required"`
	EndsAt        string  `form:"ends_at"        binding:"required"`
}

type updateEventRequest struct {
	Title         string  `form:"title"          binding:"required"`
	TitleKK       *string `form:"title_kk"`
	Description   *string `form:"description"`
	DescriptionKK *string `form:"description_kk"`
	Location      *string `form:"location"`
	StartsAt      string  `form:"starts_at"      binding:"required"`
	EndsAt        string  `form:"ends_at"        binding:"required"`
}

type eventResponse struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	TitleKK       string  `json:"title_kk"`
	Description   *string `json:"description,omitempty"`
	DescriptionKK *string `json:"description_kk,omitempty"`
	CoverURL      *string `json:"cover_url,omitempty"`
	Location      *string `json:"location,omitempty"`
	StartsAt      string  `json:"starts_at"`
	EndsAt        string  `json:"ends_at"`
	CreatedAt     string  `json:"created_at"`
}

func toEventResponse(e *entity.Event) eventResponse {
	return eventResponse{
		ID:            e.ID.String(),
		Title:         e.Title,
		TitleKK:       e.TitleKK,
		Description:   e.Description,
		DescriptionKK: e.DescriptionKK,
		CoverURL:      e.CoverURL,
		Location:      e.Location,
		StartsAt:      e.StartsAt.Format("2006-01-02T15:04:05Z"),
		EndsAt:        e.EndsAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
