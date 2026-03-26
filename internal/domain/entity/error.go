package entity

import "errors"

// Sentinel errors — используются для типизированной обработки ошибок
// в usecase и handler без парсинга строк.
var (
	// ErrNoAvailableCopies — все копии книги заняты, бронь невозможна.
	// Handler должен вернуть 409 Conflict.
	ErrNoAvailableCopies = errors.New("no available copies")

	// ErrReservationNotFound — бронь не найдена.
	ErrReservationNotFound = errors.New("reservation not found")

	// ErrForbidden — попытка выполнить действие над чужим ресурсом.
	// Handler должен вернуть 403 Forbidden.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidStatusTransition — переход из текущего статуса недопустим.
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrDuplicateReservation - если пользователь попытается забронировать одну книгу дважды
	ErrDuplicateReservation = errors.New("active reservation for this book already exists") // ДОБАВИТЬ

	ErrNotFound = errors.New("entity not found")
)
