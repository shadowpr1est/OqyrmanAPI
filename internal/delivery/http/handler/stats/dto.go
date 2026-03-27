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
