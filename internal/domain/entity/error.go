package entity

import "errors"

// Sentinel errors — используются для типизированной обработки ошибок
// в usecase и handler без парсинга строк.
var (
	// ErrValidation — ошибка валидации входных данных (пароль, формат и т.д.).
	// Handler должен вернуть 400 Bad Request с сообщением из err.Error().
	ErrValidation = errors.New("validation error")

	// ErrTokenExpired — refresh token истёк.
	// Handler должен вернуть 401 с кодом token_expired.
	ErrTokenExpired = errors.New("refresh token expired")

	// ErrAccountLocked — аккаунт временно заблокирован из-за слишком многих неудачных попыток входа.
	// Handler должен вернуть 429 Too Many Requests.
	ErrAccountLocked = errors.New("account temporarily locked due to too many failed login attempts")

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

	// ErrDuplicateReservation — попытка забронировать одну книгу дважды.
	ErrDuplicateReservation = errors.New("active reservation for this book already exists")

	// ErrDuplicateWishlist — книга уже в вишлисте пользователя.
	// Handler должен вернуть 409 Conflict.
	ErrDuplicateWishlist = errors.New("book already in wishlist")

	ErrNotFound = errors.New("entity not found")

	// ErrFileLimitExceeded — попытка загрузить больше одного файла одного типа на книгу.
	// Максимум: 1 аудио-файл (MP3) и 1 документ (PDF/EPUB) на книгу.
	// Handler должен вернуть 409 Conflict.
	ErrFileLimitExceeded = errors.New("file limit exceeded for this book")

	// ErrEmailTaken — email уже занят верифицированным пользователем.
	// Handler должен вернуть 409 Conflict.
	ErrEmailTaken = errors.New("email already taken")

	// ErrPhoneTaken — телефон уже занят верифицированным пользователем.
	// Handler должен вернуть 409 Conflict.
	ErrPhoneTaken = errors.New("phone already taken")

	// ErrActiveReservationsExist — нельзя удалить запись библиотечной книги,
	// пока есть активные или ожидающие резервации.
	// Handler должен вернуть 409 Conflict.
	ErrActiveReservationsExist = errors.New("active or pending reservations exist for this library book")
)
