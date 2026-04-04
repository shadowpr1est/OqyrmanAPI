package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type userRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *userRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := `
		INSERT INTO users (id, email, phone, password_hash, name, surname, avatar_url, role, library_id, qr_code, google_id, created_at)
		VALUES (:id, :email, :phone, :password_hash, :name, :surname, :avatar_url, :role, :library_id, :qr_code, :google_id, :created_at)
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("userRepo.Create: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("userRepo.Create rows error: %w", err)
		}
		return nil, fmt.Errorf("userRepo.Create: no rows returned")
	}
	if err := rows.StructScan(user); err != nil {
		return nil, fmt.Errorf("userRepo.Create scan: %w", err)
	}
	return user, nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("userRepo.GetByID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`, email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("userRepo.GetByEmail: %w", err)
	}
	return &user, nil
}

func (r *userRepo) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	// role, library_id, qr_code, avatar_url намеренно исключены — не меняются через профиль.
	// CASE WHEN: пустая строка = поле не передано → сохраняем текущее значение в БД.
	// Это позволяет хэндлеру не делать предварительный GetByID для partial update.
	query := `
		UPDATE users
		SET email    = CASE WHEN :email   <> '' THEN :email   ELSE email   END,
		    phone    = CASE WHEN :phone   <> '' THEN :phone   ELSE phone   END,
		    name     = CASE WHEN :name    <> '' THEN :name    ELSE name    END,
		    surname  = CASE WHEN :surname <> '' THEN :surname ELSE surname END,
		WHERE id = :id AND deleted_at IS NULL
		RETURNING *`
	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, entity.ErrEmailTaken
		}
		return nil, fmt.Errorf("userRepo.Update: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("userRepo.Update rows error: %w", err)
		}
		return nil, entity.ErrNotFound
	}
	if err := rows.StructScan(user); err != nil {
		return nil, fmt.Errorf("userRepo.Update scan: %w", err)
	}
	return user, nil
}

// Delete — soft delete.
// Токены не трогаем — они протухнут сами по TTL.
// При следующем запросе GetByEmail/GetByID вернёт ErrNotFound.
func (r *userRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return fmt.Errorf("userRepo.Delete: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("userRepo.Delete rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *userRepo) GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	var user entity.User
	err := r.db.GetContext(ctx, &user,
		`SELECT * FROM users WHERE google_id = $1 AND deleted_at IS NULL`, googleID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}
		return nil, fmt.Errorf("userRepo.GetByGoogleID: %w", err)
	}
	return &user, nil
}

func (r *userRepo) SetGoogleID(ctx context.Context, id uuid.UUID, googleID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET google_id = $1 WHERE id = $2`, googleID, id,
	)
	return err
}

func (r *userRepo) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email_verified = true, email_verified_at = now() WHERE id = $1`, id,
	)
	return err
}

func (r *userRepo) UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE users SET avatar_url = $1 WHERE id = $2 AND deleted_at IS NULL`,
		avatarURL, id,
	)
	if err != nil {
		return fmt.Errorf("userRepo.UpdateAvatarURL: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("userRepo.UpdateAvatarURL rows affected: %w", err)
	}
	if rows == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *userRepo) ListAllView(ctx context.Context, limit, offset int) ([]*entity.UserView, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total,
		`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`,
	); err != nil {
		return nil, 0, fmt.Errorf("userRepo.ListAllView count: %w", err)
	}
	var users []*entity.UserView
	if err := r.db.SelectContext(ctx, &users, `
		SELECT u.id, u.email, u.name, u.surname, u.phone,
		       COALESCE(u.avatar_url, '') AS avatar_url,
		       u.role, u.library_id, COALESCE(l.name, '') AS library_name,
		       u.qr_code, u.created_at
		FROM users u
		LEFT JOIN libraries l ON l.id = u.library_id
		WHERE u.deleted_at IS NULL
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2`,
		limit, offset,
	); err != nil {
		return nil, 0, fmt.Errorf("userRepo.ListAllView: %w", err)
	}
	return users, total, nil
}
func (r *userRepo) AdminUpdate(
	ctx context.Context,
	id uuid.UUID,
	role *entity.Role,
	libraryID *uuid.UUID,
	name, surname, email, phone *string,
) (*entity.UserView, error) {
	query := `
    WITH updated AS (
        UPDATE users
        SET role       = COALESCE($1, role),
            library_id = CASE WHEN $1 IS NOT NULL THEN $2 ELSE library_id END,
            name       = COALESCE(NULLIF($3, ''), name),
            surname    = COALESCE(NULLIF($4, ''), surname),
            email      = COALESCE(NULLIF($5, ''), email),
            phone      = COALESCE(NULLIF($6, ''), phone),
        WHERE id = $7 AND deleted_at IS NULL
        RETURNING *
    )
    SELECT 
        u.id, u.email, u.name, u.surname, u.phone,
        COALESCE(u.avatar_url, '') AS avatar_url,
        u.role, u.library_id,
        COALESCE(l.name, '') AS library_name,
        u.qr_code, u.created_at
    FROM updated u
    LEFT JOIN libraries l ON l.id = u.library_id;
    `

	var roleStr *string
	if role != nil {
		s := string(*role)
		roleStr = &s
	}
	nameStr := ""
	if name != nil {
		nameStr = *name
	}
	surnameStr := ""
	if surname != nil {
		surnameStr = *surname
	}
	emailStr := ""
	if email != nil {
		emailStr = *email
	}
	phoneStr := ""
	if phone != nil {
		phoneStr = *phone
	}

	// Возвращаем обновлённый UserView с именем библиотеки
	var view entity.UserView
	err := r.db.GetContext(ctx, &view, query,
		roleStr, libraryID, nameStr, surnameStr, emailStr, phoneStr, id,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, entity.ErrNotFound
		}

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, entity.ErrEmailTaken
		}

		return nil, fmt.Errorf("userRepo.AdminUpdate: %w", err)
	}

	return &view, nil
}
