package entity

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin Role = "Admin"
	RoleStaff Role = "Staff"
	RoleUser  Role = "User"
)

type User struct {
	ID           uuid.UUID  `db:"id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	FullName     string     `db:"full_name"`
	Phone        string     `db:"phone"`
	AvatarURL    string     `db:"avatar_url"`
	Role         Role       `db:"role"`
	LibraryID    *uuid.UUID `db:"library_id"` // ← НОВОЕ: NULL для admin/user
	QRCode       string     `db:"qr_code"`
	CreatedAt    time.Time  `db:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}
