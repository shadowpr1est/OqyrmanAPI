package notification

type notificationResponse struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	IsRead    bool    `json:"is_read"`
	CreatedAt string  `json:"created_at"`
	ReadAt    *string `json:"read_at,omitempty"`
}

type listNotificationsResponse struct {
	Items  []*notificationResponse `json:"items"`
	Total  int                     `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
}
