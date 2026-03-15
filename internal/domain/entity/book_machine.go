package entity

import "github.com/google/uuid"

type MachineStatus string

const (
	MachineStatusActive   MachineStatus = "active"
	MachineStatusInactive MachineStatus = "inactive"
)

type BookMachine struct {
	ID      uuid.UUID     `db:"id"`
	Name    string        `db:"name"`
	Address string        `db:"address"`
	Lat     float64       `db:"lat"`
	Lng     float64       `db:"lng"`
	Status  MachineStatus `db:"status"`
}
