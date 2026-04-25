package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type reservationRepo struct {
	db *sqlx.DB
}

func NewReservationRepo(db *sqlx.DB) *reservationRepo {
	return &reservationRepo{db: db}
}

func generateQRToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateQRToken: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (r *reservationRepo) CreateWithDecrement(ctx context.Context, res *entity.Reservation) (*entity.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement begin tx: %w", err)
	}
	defer tx.Rollback()

	// Проверка дубля: нельзя бронировать одну книгу дважды
	var count int
	if err := tx.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM reservations r
		JOIN library_books lb ON lb.id = r.library_book_id
		WHERE r.user_id = $1
		  AND lb.book_id = (SELECT book_id FROM library_books WHERE id = $2)
		  AND r.status IN ('pending', 'active')`,
		res.UserID, res.LibraryBookID,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement check duplicate: %w", err)
	}
	if count > 0 {
		return nil, entity.ErrDuplicateReservation
	}

	// Декремент available_copies с проверкой наличия
	var available int
	if err := tx.GetContext(ctx, &available,
		`SELECT available_copies FROM library_books WHERE id = $1 FOR UPDATE`,
		res.LibraryBookID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement get copies: %w", err)
	}
	if available <= 0 {
		return nil, entity.ErrNoAvailableCopies
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE library_books SET available_copies = available_copies - 1 WHERE id = $1`,
		res.LibraryBookID,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement decrement: %w", err)
	}

	// Генерация уникального QR-токена для бронирования
	qrToken, err := generateQRToken()
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement qr token: %w", err)
	}
	res.QRToken = qrToken

	query := `
		INSERT INTO reservations (id, user_id, library_book_id, status, reserved_at, due_date, returned_at, extended_count, qr_token)
		VALUES (:id, :user_id, :library_book_id, :status, :reserved_at, :due_date, :returned_at, :extended_count, :qr_token)
		RETURNING *`

	rows, err := sqlx.NamedQueryContext(ctx, tx, query, res)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement insert: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement rows error: %w", err)
		}
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement: no rows returned")
	}
	if err := rows.StructScan(res); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement scan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement commit: %w", err)
	}

	return res, nil
}

// closeAndIncrement — общая логика для cancel и return.
// callerID != nil → проверяет что бронь принадлежит пользователю
// libraryID != nil → проверяет что бронь принадлежит библиотеке staff
// оба nil → admin, без проверок
func (r *reservationRepo) closeAndIncrement(
	ctx context.Context,
	id uuid.UUID,
	callerID *uuid.UUID,
	libraryID *uuid.UUID,
	targetStatus entity.ReservationStatus,
	allowedStatuses []entity.ReservationStatus,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement begin tx: %w", err)
	}
	defer tx.Rollback()

	var res entity.Reservation
	if err := tx.GetContext(ctx, &res,
		`SELECT * FROM reservations WHERE id = $1 FOR UPDATE`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrReservationNotFound
		}
		return fmt.Errorf("reservationRepo.closeAndIncrement get: %w", err)
	}

	// Проверка владельца — для отмены пользователем
	if callerID != nil && res.UserID != *callerID {
		return fmt.Errorf("%w: reservation belongs to another user", entity.ErrForbidden)
	}

	// Проверка принадлежности библиотеке — для staff операций
	if libraryID != nil {
		var belongs bool
		if err := tx.GetContext(ctx, &belongs, `
			SELECT EXISTS(
				SELECT 1 FROM library_books lb
				WHERE lb.id = $1 AND lb.library_id = $2
			)`, res.LibraryBookID, *libraryID,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement check library: %w", err)
		}
		if !belongs {
			return fmt.Errorf("%w: reservation belongs to another library", entity.ErrForbidden)
		}
	}

	if !containsStatus(allowedStatuses, res.Status) {
		return fmt.Errorf("%w: from '%s' to '%s'",
			entity.ErrInvalidStatusTransition, res.Status, targetStatus)
	}

	if targetStatus == entity.ReservationCompleted {
		if _, err := tx.ExecContext(ctx,
			`UPDATE reservations SET status = $1, returned_at = now() WHERE id = $2`,
			targetStatus, id,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement update status: %w", err)
		}
	} else {
		if _, err := tx.ExecContext(ctx,
			`UPDATE reservations SET status = $1 WHERE id = $2`,
			targetStatus, id,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement update status: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE library_books SET available_copies = available_copies + 1 WHERE id = $1`,
		res.LibraryBookID,
	); err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement increment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement commit: %w", err)
	}

	return nil
}

func (r *reservationRepo) CancelWithIncrement(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error {
	return r.closeAndIncrement(ctx, id, &callerID, nil,
		entity.ReservationCancelled,
		[]entity.ReservationStatus{entity.ReservationPending},
	)
}

func (r *reservationRepo) ListByLibrary(
	ctx context.Context,
	libraryID uuid.UUID,
	limit, offset int,
	status *string,
) ([]*entity.Reservation, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(res.id)
		FROM reservations res
		JOIN library_books lb ON lb.id = res.library_book_id
		WHERE lb.library_id = $1
		  AND ($2::text IS NULL OR res.status = $2)`,
		libraryID, status,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByLibrary count: %w", err)
	}

	var items []*entity.Reservation
	if err := r.db.SelectContext(ctx, &items, `
		SELECT res.*
		FROM reservations res
		JOIN library_books lb ON lb.id = res.library_book_id
		WHERE lb.library_id = $1
		  AND ($2::text IS NULL OR res.status = $2)
		ORDER BY res.reserved_at DESC
		LIMIT $3 OFFSET $4`,
		libraryID, status, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByLibrary: %w", err)
	}

	return items, total, nil
}

func (r *reservationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	var res entity.Reservation
	if err := r.db.GetContext(ctx, &res,
		`SELECT * FROM reservations WHERE id = $1`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, fmt.Errorf("reservationRepo.GetByID: %w", err)
	}
	return &res, nil
}

func (r *reservationRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Reservation, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM reservations WHERE user_id = $1`, userID,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByUser count: %w", err)
	}

	var items []*entity.Reservation
	if err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM reservations
		WHERE user_id = $1
		ORDER BY reserved_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByUser: %w", err)
	}

	return items, total, nil
}

func (r *reservationRepo) ListAll(ctx context.Context, limit, offset int, status *string) ([]*entity.Reservation, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM reservations WHERE ($1::text IS NULL OR status = $1)`, status,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll count: %w", err)
	}

	var items []*entity.Reservation
	if err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM reservations
		WHERE ($1::text IS NULL OR status = $1)
		ORDER BY reserved_at DESC
		LIMIT $2 OFFSET $3`,
		status, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll: %w", err)
	}

	return items, total, nil
}

func (r *reservationRepo) Extend(ctx context.Context, id, userID uuid.UUID) (*entity.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.Extend begin tx: %w", err)
	}
	defer tx.Rollback()

	var res entity.Reservation
	if err := tx.GetContext(ctx, &res,
		`SELECT * FROM reservations WHERE id = $1 FOR UPDATE`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, fmt.Errorf("reservationRepo.Extend get: %w", err)
	}

	if res.UserID != userID {
		return nil, fmt.Errorf("%w: reservation belongs to another user", entity.ErrForbidden)
	}
	if res.Status != entity.ReservationActive {
		return nil, fmt.Errorf("%w: can only extend active reservations", entity.ErrInvalidStatusTransition)
	}
	if res.ExtendedCount >= 1 {
		return nil, entity.ErrExtendLimitReached
	}

	// Продление на 7 дней от текущего дедлайна
	newDueDate := res.DueDate.AddDate(0, 0, 7)

	if err := tx.GetContext(ctx, &res, `
		UPDATE reservations
		SET due_date = $2, extended_count = extended_count + 1
		WHERE id = $1
		RETURNING *`,
		id, newDueDate,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.Extend update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reservationRepo.Extend commit: %w", err)
	}

	return &res, nil
}

// UpdateStatus is the admin variant: no library-ownership check.
// Terminal transitions (cancelled/completed) atomically restore available_copies.
// pending → active records the physical handout; copies were already decremented on creation.
func (r *reservationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	switch status {
	case entity.ReservationCancelled:
		return r.closeAndIncrement(ctx, id, nil, nil,
			entity.ReservationCancelled,
			[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
		)
	case entity.ReservationCompleted:
		return r.closeAndIncrement(ctx, id, nil, nil,
			entity.ReservationCompleted,
			[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
		)
	case entity.ReservationActive:
		return r.activateReservation(ctx, id, nil)
	default:
		return fmt.Errorf("%w: admin cannot transition to '%s'",
			entity.ErrInvalidStatusTransition, status)
	}
}

func (r *reservationRepo) CancelOverdue(ctx context.Context) ([]entity.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CancelOverdue begin tx: %w", err)
	}
	defer tx.Rollback()

	// Lock and collect IDs in one query.
	var overdue []entity.Reservation
	if err := tx.SelectContext(ctx, &overdue, `
		SELECT * FROM reservations
		WHERE due_date < now() AND status IN ('active', 'pending')
		FOR UPDATE`,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.CancelOverdue select: %w", err)
	}
	if len(overdue) == 0 {
		return nil, nil
	}

	ids := make([]uuid.UUID, len(overdue))
	for i, res := range overdue {
		ids[i] = res.ID
	}

	// Batch-cancel all overdue reservations in one UPDATE.
	if _, err := tx.ExecContext(ctx,
		`UPDATE reservations SET status = 'cancelled' WHERE id = ANY($1)`,
		pq.Array(ids),
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.CancelOverdue batch update: %w", err)
	}

	// Batch-restore available_copies: group by library_book_id, increment by count.
	if _, err := tx.ExecContext(ctx, `
		UPDATE library_books lb
		SET available_copies = lb.available_copies + sub.cnt
		FROM (
			SELECT library_book_id, COUNT(*) AS cnt
			FROM reservations
			WHERE id = ANY($1)
			GROUP BY library_book_id
		) sub
		WHERE lb.id = sub.library_book_id`,
		pq.Array(ids),
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.CancelOverdue batch increment: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reservationRepo.CancelOverdue commit: %w", err)
	}

	return overdue, nil
}

func (r *reservationRepo) FindApproachingDeadline(ctx context.Context, within time.Duration) ([]entity.Reservation, error) {
	var items []entity.Reservation
	if err := r.db.SelectContext(ctx, &items, `
		SELECT * FROM reservations
		WHERE status IN ('active', 'pending')
		  AND due_date > now()
		  AND due_date <= now() + $1::interval
		ORDER BY due_date ASC`,
		within.String(),
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.FindApproachingDeadline: %w", err)
	}
	return items, nil
}

func (r *reservationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx,
		`DELETE FROM reservations WHERE id = $1`, id,
	); err != nil {
		return fmt.Errorf("reservationRepo.Delete: %w", err)
	}
	return nil
}

func (r *reservationRepo) StaffCancel(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	return r.closeAndIncrement(ctx, id, nil, &libraryID,
		entity.ReservationCancelled,
		[]entity.ReservationStatus{entity.ReservationPending},
	)
}

func (r *reservationRepo) StaffReturn(ctx context.Context, id uuid.UUID, libraryID uuid.UUID) error {
	return r.closeAndIncrement(ctx, id, nil, &libraryID,
		entity.ReservationCompleted,
		[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
	)
}

func (r *reservationRepo) AdminReturn(ctx context.Context, id uuid.UUID) error {
	return r.closeAndIncrement(ctx, id, nil, nil,
		entity.ReservationCompleted,
		[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
	)
}

// StaffUpdateStatus is the staff variant: enforces library-ownership on every transition.
// Terminal transitions (cancelled/completed) atomically restore available_copies.
// pending → active records the physical handout; copies were already decremented on creation.
func (r *reservationRepo) StaffUpdateStatus(ctx context.Context, id uuid.UUID, libraryID uuid.UUID, status entity.ReservationStatus) error {
	switch status {
	case entity.ReservationCancelled:
		return r.closeAndIncrement(ctx, id, nil, &libraryID,
			entity.ReservationCancelled,
			[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
		)
	case entity.ReservationCompleted:
		return r.closeAndIncrement(ctx, id, nil, &libraryID,
			entity.ReservationCompleted,
			[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
		)
	case entity.ReservationActive:
		return r.activateReservation(ctx, id, &libraryID)
	default:
		return fmt.Errorf("%w: staff cannot transition to '%s'",
			entity.ErrInvalidStatusTransition, status)
	}
}

// activateReservation handles the pending → active transition (book physically handed out).
// available_copies is NOT changed — it was already decremented on reservation creation.
// If libraryID is non-nil, the reservation must belong to that library (staff context).
func (r *reservationRepo) activateReservation(ctx context.Context, id uuid.UUID, libraryID *uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("reservationRepo.activateReservation begin tx: %w", err)
	}
	defer tx.Rollback()

	var res entity.Reservation
	if err := tx.GetContext(ctx, &res,
		`SELECT * FROM reservations WHERE id = $1 FOR UPDATE`, id,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ErrReservationNotFound
		}
		return fmt.Errorf("reservationRepo.activateReservation get: %w", err)
	}

	if libraryID != nil {
		var belongs bool
		if err := tx.GetContext(ctx, &belongs, `
			SELECT EXISTS(
				SELECT 1 FROM library_books lb
				WHERE lb.id = $1 AND lb.library_id = $2
			)`, res.LibraryBookID, *libraryID,
		); err != nil {
			return fmt.Errorf("reservationRepo.activateReservation check library: %w", err)
		}
		if !belongs {
			return fmt.Errorf("%w: reservation not found in this library", entity.ErrForbidden)
		}
	}

	if res.Status != entity.ReservationPending {
		return fmt.Errorf("%w: from '%s' to 'active'",
			entity.ErrInvalidStatusTransition, res.Status)
	}

	// При активации книга выдаётся на 30 дней
	newDueDate := time.Now().AddDate(0, 0, 30)
	if _, err := tx.ExecContext(ctx,
		`UPDATE reservations SET status = 'active', due_date = $2 WHERE id = $1`, id, newDueDate,
	); err != nil {
		return fmt.Errorf("reservationRepo.activateReservation update: %w", err)
	}

	return tx.Commit()
}

// reservationViewQuery is the base SELECT for all ReservationView methods, ДЛЯ АДМИНОВ И СТАФФ, НЕ ДЛЯ ОБЫЧНЫХ ПОЛЬЗОВАТЕЛЕЙ
const reservationViewQuery = `
	SELECT res.id, res.status, res.reserved_at, res.due_date, res.returned_at,
	       res.extended_count, res.qr_token,
	       res.user_id, u.name AS user_name, u.surname AS user_surname, u.email AS user_email,
	       res.library_book_id, b.id AS book_id, b.title AS book_title,
	       COALESCE(b.cover_url, '') AS book_cover_url,
	       lb.library_id, l.name AS library_name
	FROM reservations res
	JOIN users         u  ON u.id  = res.user_id          AND u.deleted_at  IS NULL
	JOIN library_books lb ON lb.id = res.library_book_id
	JOIN books         b  ON b.id  = lb.book_id           AND b.deleted_at  IS NULL
	JOIN libraries     l  ON l.id  = lb.library_id        AND l.deleted_at  IS NULL`

func (r *reservationRepo) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReservationView, error) {
	var v entity.ReservationView
	err := r.db.GetContext(ctx, &v,
		reservationViewQuery+` WHERE res.id = $1`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, fmt.Errorf("reservationRepo.GetByIDView: %w", err)
	}
	return &v, nil
}

func (r *reservationRepo) ListByUserView(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.ReservationView, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM reservations WHERE user_id = $1`, userID,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByUserView count: %w", err)
	}
	var items []*entity.ReservationView
	if err := r.db.SelectContext(ctx, &items,
		reservationViewQuery+` WHERE res.user_id = $1 ORDER BY res.reserved_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByUserView: %w", err)
	}
	return items, total, nil
}

func (r *reservationRepo) ListByLibraryView(ctx context.Context, libraryID uuid.UUID, limit, offset int, status *string) ([]*entity.ReservationView, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(res.id) FROM reservations res
		JOIN library_books lb ON lb.id = res.library_book_id
		WHERE lb.library_id = $1 AND ($2::text IS NULL OR res.status = $2)`,
		libraryID, status,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByLibraryView count: %w", err)
	}
	var items []*entity.ReservationView
	if err := r.db.SelectContext(ctx, &items,
		reservationViewQuery+` WHERE lb.library_id = $1 AND ($2::text IS NULL OR res.status = $2)
		ORDER BY res.reserved_at DESC LIMIT $3 OFFSET $4`,
		libraryID, status, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListByLibraryView: %w", err)
	}
	return items, total, nil
}

// Это для админа так что это ок, если он может видеть данные пользователя.
func (r *reservationRepo) ListAllView(ctx context.Context, limit, offset int, status *string) ([]*entity.ReservationView, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM reservations WHERE ($1::text IS NULL OR status = $1)`, status,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAllView count: %w", err)
	}
	var items []*entity.ReservationView
	if err := r.db.SelectContext(ctx, &items,
		reservationViewQuery+` WHERE ($1::text IS NULL OR res.status = $1) ORDER BY res.reserved_at DESC LIMIT $2 OFFSET $3`,
		status, limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAllView: %w", err)
	}
	return items, total, nil
}

// ActivateByQRToken atomically looks up a reservation by its QR token,
// verifies it belongs to the staff's library and is pending, then activates it.
func (r *reservationRepo) ActivateByQRToken(ctx context.Context, qrToken string, libraryID uuid.UUID) (*entity.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.ActivateByQRToken begin tx: %w", err)
	}
	defer tx.Rollback()

	var res entity.Reservation
	if err := tx.GetContext(ctx, &res,
		`SELECT * FROM reservations WHERE qr_token = $1 FOR UPDATE`, qrToken,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrReservationNotFound
		}
		return nil, fmt.Errorf("reservationRepo.ActivateByQRToken get: %w", err)
	}

	// Проверка принадлежности библиотеке
	var belongs bool
	if err := tx.GetContext(ctx, &belongs, `
		SELECT EXISTS(
			SELECT 1 FROM library_books lb
			WHERE lb.id = $1 AND lb.library_id = $2
		)`, res.LibraryBookID, libraryID,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.ActivateByQRToken check library: %w", err)
	}
	if !belongs {
		return nil, fmt.Errorf("%w: reservation not found in this library", entity.ErrForbidden)
	}

	if res.Status != entity.ReservationPending {
		return nil, fmt.Errorf("%w: from '%s' to 'active'",
			entity.ErrInvalidStatusTransition, res.Status)
	}

	// Активация: статус active, срок 30 дней
	newDueDate := time.Now().AddDate(0, 0, 30)
	if err := tx.GetContext(ctx, &res, `
		UPDATE reservations SET status = 'active', due_date = $2
		WHERE id = $1
		RETURNING *`, res.ID, newDueDate,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.ActivateByQRToken update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reservationRepo.ActivateByQRToken commit: %w", err)
	}

	return &res, nil
}

func (r *reservationRepo) ListPendingByUserAndLibraryView(ctx context.Context, userID, libraryID uuid.UUID) ([]*entity.ReservationView, error) {
	var items []*entity.ReservationView
	if err := r.db.SelectContext(ctx, &items,
		reservationViewQuery+` WHERE res.user_id = $1 AND lb.library_id = $2 AND res.status = 'pending'
		ORDER BY res.reserved_at DESC`,
		userID, libraryID,
	); err != nil {
		return nil, fmt.Errorf("reservationRepo.ListPendingByUserAndLibraryView: %w", err)
	}
	return items, nil
}

func containsStatus(allowed []entity.ReservationStatus, current entity.ReservationStatus) bool {
	for _, s := range allowed {
		if s == current {
			return true
		}
	}
	return false
}
