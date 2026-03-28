package entity

type Stats struct {
	UsersTotal          int `db:"users_total"`
	BooksTotal          int `db:"books_total"`
	AuthorsTotal        int `db:"authors_total"`
	ReservationsActive  int `db:"reservations_active"`
	ReservationsPending int `db:"reservations_pending"`
	ReservationsTotal   int `db:"reservations_total"`
	ReviewsTotal        int `db:"reviews_total"`
}

type UserStats struct {
	BooksRead          int `db:"books_read"           json:"books_read"`
	ActiveReservations int `db:"active_reservations"  json:"active_reservations"`
	ReviewsGiven       int `db:"reviews_given"        json:"reviews_given"`
	WishlistCount      int `db:"wishlist_count"       json:"wishlist_count"`
}

type LibraryStats struct {
	TotalBooks            int `db:"total_books"            json:"total_books"`
	AvailableBooks        int `db:"available_books"        json:"available_books"`
	TotalReservations     int `db:"total_reservations"     json:"total_reservations"`
	ActiveReservations    int `db:"active_reservations"    json:"active_reservations"`
	PendingReservations   int `db:"pending_reservations"   json:"pending_reservations"`
	CompletedReservations int `db:"completed_reservations" json:"completed_reservations"`
	CancelledReservations int `db:"cancelled_reservations" json:"cancelled_reservations"`
}
