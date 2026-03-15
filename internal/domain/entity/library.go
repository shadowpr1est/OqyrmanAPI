package entity

import "github.com/google/uuid"

type Library struct {
	ID      uuid.UUID `db:"id"`
	Name    string    `db:"name"`
	Address string    `db:"address"`
	Lat     float64   `db:"lat"`
	Lng     float64   `db:"lng"`
	Phone   string    `db:"phone"`
}
