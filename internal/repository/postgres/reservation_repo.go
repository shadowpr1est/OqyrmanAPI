package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type reservationRepo struct {
	db *sqlx.DB
}

func NewReservationRepo(db *sqlx.DB) *reservationRepo {
	return &reservationRepo{db: db}
}

// CreateWithDecrement атомарно:
// 1. Блокирует строку копий (FOR UPDATE) — исключает race condition
// 2. Проверяет available_copies > 0
// 3. Уменьшает available_copies на 1
// 4. Создаёт запись бронирования
// Всё в одной транзакции — либо всё, либо ничего.
func (r *reservationRepo) CreateWithDecrement(ctx context.Context, res *entity.Reservation) (*entity.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement begin tx: %w", err)
	}
	// Откатываем транзакцию при любой ошибке.
	// После успешного Commit() этот вызов — no-op.
	defer tx.Rollback()

	if res.LibraryBookID != nil {
		var available int
		// FOR UPDATE блокирует строку до конца транзакции —
		// параллельный запрос будет ждать, а не читать устаревшие данные.
		err := tx.GetContext(ctx, &available,
			`SELECT available_copies FROM library_books WHERE id = $1 FOR UPDATE`,
			*res.LibraryBookID,
		)
		if err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement get library_book: %w", err)
		}
		if available <= 0 {
			return nil, fmt.Errorf("no available copies in library")
		}
		if _, err = tx.ExecContext(ctx,
			`UPDATE library_books SET available_copies = available_copies - 1 WHERE id = $1`,
			*res.LibraryBookID,
		); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement decrement library: %w", err)
		}
	}

	if res.MachineBookID != nil {
		var available int
		err := tx.GetContext(ctx, &available,
			`SELECT available_copies FROM book_machine_books WHERE id = $1 FOR UPDATE`,
			*res.MachineBookID,
		)
		if err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement get machine_book: %w", err)
		}
		if available <= 0 {
			return nil, fmt.Errorf("no available copies in book machine")
		}
		if _, err = tx.ExecContext(ctx,
			`UPDATE book_machine_books SET available_copies = available_copies - 1 WHERE id = $1`,
			*res.MachineBookID,
		); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement decrement machine: %w", err)
		}
	}

	// Вставляем бронь — используем именованный запрос через sqlx.
	// tx.NamedQuery работает так же, как db.NamedQueryContext, но в транзакции.
	query := `
		INSERT INTO reservations
			(id, user_id, library_book_id, machine_book_id, source_type, status, reserved_at, due_date, returned_at)
		VALUES
			(:id, :user_id, :library_book_id, :machine_book_id, :source_type, :status, :reserved_at, :due_date, :returned_at)
		RETURNING *`

	rows, err := tx.NamedQuery(query, res)
	if err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement insert: %w", err)
	}
	defer rows.Close()

	// FIX ПРИОРИТЕТ 4: проверяем rows.Next() — без этого StructScan паникует
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("reservationRepo.CreateWithDecrement rows error: %w", err)
		}
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement: no rows returned after insert")
	}
	if err := rows.StructScan(res); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement scan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("reservationRepo.CreateWithDecrement commit: %w", err)
	}

	return res, nil
}

// closeAndIncrement — приватный хелпер для атомарного закрытия брони.
// Выполняет в одной транзакции:
//  1. SELECT FOR UPDATE — блокирует строку, читает актуальный статус
//  2. Валидация — проверяет что переход из текущего статуса допустим
//  3. UPDATE reservations — меняет статус; returned_at проставляется
//     автоматически только для completed
//  4. UPDATE copies — увеличивает available_copies на 1
//
// Статус не используется как bool-флаг — returned_at выводится из него:
// completed → returned_at = now(), cancelled → returned_at остаётся NULL.
func (r *reservationRepo) closeAndIncrement(
	ctx context.Context,
	id uuid.UUID,
	targetStatus entity.ReservationStatus,
	allowedStatuses []entity.ReservationStatus,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement begin tx: %w", err)
	}
	defer tx.Rollback()

	// FOR UPDATE блокирует строку — параллельный запрос будет ждать
	var res entity.Reservation
	if err := tx.GetContext(ctx, &res,
		`SELECT * FROM reservations WHERE id = $1 FOR UPDATE`, id,
	); err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement get reservation: %w", err)
	}

	// Валидация перехода статуса — защита от двойного возврата/отмены
	if !containsStatus(allowedStatuses, res.Status) {
		return fmt.Errorf("cannot transition reservation from '%s' to '%s'",
			res.Status, targetStatus)
	}

	// returned_at проставляем только при completed — выводим из статуса,
	// не передаём как bool-параметр
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

	// Возвращаем копию в наличие
	if res.LibraryBookID != nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE library_books SET available_copies = available_copies + 1 WHERE id = $1`,
			*res.LibraryBookID,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement increment library: %w", err)
		}
	} else if res.MachineBookID != nil {
		if _, err := tx.ExecContext(ctx,
			`UPDATE book_machine_books SET available_copies = available_copies + 1 WHERE id = $1`,
			*res.MachineBookID,
		); err != nil {
			return fmt.Errorf("reservationRepo.closeAndIncrement increment machine: %w", err)
		}
	} else {
		return fmt.Errorf("reservation %s has no associated book copy", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("reservationRepo.closeAndIncrement commit: %w", err)
	}

	return nil
}

// containsStatus — вспомогательная функция для проверки допустимости перехода
func containsStatus(allowed []entity.ReservationStatus, current entity.ReservationStatus) bool {
	for _, s := range allowed {
		if s == current {
			return true
		}
	}
	return false
}

// ReturnWithIncrement — пользователь вернул физическую книгу.
// Допустимые исходные статусы: pending, active.
func (r *reservationRepo) ReturnWithIncrement(ctx context.Context, id uuid.UUID) error {
	return r.closeAndIncrement(ctx, id,
		entity.ReservationCompleted,
		[]entity.ReservationStatus{entity.ReservationPending, entity.ReservationActive},
	)
}

// CancelWithIncrement — пользователь отменил бронь до получения книги.
// Допустимый исходный статус: только pending.
func (r *reservationRepo) CancelWithIncrement(ctx context.Context, id uuid.UUID) error {
	return r.closeAndIncrement(ctx, id,
		entity.ReservationCancelled,
		[]entity.ReservationStatus{entity.ReservationPending},
	)
}

func (r *reservationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Reservation, error) {
	var res entity.Reservation
	query := `SELECT * FROM reservations WHERE id = $1`
	if err := r.db.GetContext(ctx, &res, query, id); err != nil {
		return nil, fmt.Errorf("reservationRepo.GetByID: %w", err)
	}
	return &res, nil
}

func (r *reservationRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Reservation, error) {
	var items []*entity.Reservation
	query := `SELECT * FROM reservations WHERE user_id = $1 ORDER BY reserved_at DESC`
	if err := r.db.SelectContext(ctx, &items, query, userID); err != nil {
		return nil, fmt.Errorf("reservationRepo.ListByUser: %w", err)
	}
	return items, nil
}

func (r *reservationRepo) ListAll(ctx context.Context, limit, offset int) ([]*entity.Reservation, int, error) {
	var items []*entity.Reservation
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM reservations`); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll count: %w", err)
	}
	query := `SELECT * FROM reservations ORDER BY reserved_at DESC LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &items, query, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("reservationRepo.ListAll: %w", err)
	}
	return items, total, nil
}

func (r *reservationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ReservationStatus) error {
	var err error
	if status == entity.ReservationCompleted {
		_, err = r.db.ExecContext(ctx,
			`UPDATE reservations SET status = $1, returned_at = now() WHERE id = $2`, status, id)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE reservations SET status = $1 WHERE id = $2`, status, id)
	}
	if err != nil {
		return fmt.Errorf("reservationRepo.UpdateStatus: %w", err)
	}
	return nil
}

func (r *reservationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM reservations WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		return fmt.Errorf("reservationRepo.Delete: %w", err)
	}
	return nil
}
