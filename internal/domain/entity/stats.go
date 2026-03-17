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
