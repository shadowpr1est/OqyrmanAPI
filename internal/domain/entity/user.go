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
	Name         string     `db:"name"`
	Surname      string     `db:"surname"`
	FullName     string     `db:"full_name"`
	Phone        string     `db:"phone"`
	AvatarURL    string     `db:"avatar_url"`
	Role         Role       `db:"role"`
	LibraryID    *uuid.UUID `db:"library_id"`
	QRCode       string     `db:"qr_code"`
	CreatedAt    time.Time  `db:"created_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

// UserView — read model for GET /admin/users (includes library name via LEFT JOIN).
type UserView struct {
	ID          uuid.UUID  `db:"id"`
	Email       string     `db:"email"`
	Name        string     `db:"name"`
	Surname     string     `db:"surname"`
	FullName    string     `db:"full_name"`
	Phone       string     `db:"phone"`
	AvatarURL   string     `db:"avatar_url"`
	Role        Role       `db:"role"`
	LibraryID   *uuid.UUID `db:"library_id"`
	LibraryName string     `db:"library_name"`
	QRCode      string     `db:"qr_code"`
	CreatedAt   time.Time  `db:"created_at"`
}
