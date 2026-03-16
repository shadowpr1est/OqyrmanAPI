package ai

type chatRequest struct {
	Message string `json:"message" binding:"required"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

type recommendResponse struct {
	Recommendations string `json:"recommendations"`
}
