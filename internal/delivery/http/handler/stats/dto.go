package stats

type statsResponse struct {
	UsersTotal          int `json:"users_total"`
	BooksTotal          int `json:"books_total"`
	AuthorsTotal        int `json:"authors_total"`
	ReservationsActive  int `json:"reservations_active"`
	ReservationsPending int `json:"reservations_pending"`
	ReservationsTotal   int `json:"reservations_total"`
	ReviewsTotal        int `json:"reviews_total"`
}

type userStatsResponse struct {
	BooksRead          int `json:"books_read"`
	ActiveReservations int `json:"active_reservations"`
	ReviewsGiven       int `json:"reviews_given"`
	WishlistCount      int `json:"wishlist_count"`
}

type libraryStatsResponse struct {
	TotalBooks            int `json:"total_books"`
	AvailableBooks        int `json:"available_books"`
	TotalReservations     int `json:"total_reservations"`
	ActiveReservations    int `json:"active_reservations"`
	PendingReservations   int `json:"pending_reservations"`
	CompletedReservations int `json:"completed_reservations"`
	CancelledReservations int `json:"cancelled_reservations"`
}
