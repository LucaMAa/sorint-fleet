package dto

import (
	"errors"
	"sorint-fleet/internal/model"

	"github.com/google/uuid"
)

type CreateVehicleDto struct {
	LicensePlate  string `json:"license_plate" binding:"required"`
	Brand         string `json:"brand"         binding:"required"`
	Model         string `json:"model"         binding:"required"`
	Year          int    `json:"year"          binding:"required,min=1900"`
	Color         string `json:"color"`
	FuelType      string `json:"fuel_type"`
	Mileage       int    `json:"mileage"`
	Notes         string `json:"notes"`
	Jolly         bool   `json:"jolly"`
	JollyDuration int    `json:"jolly_duration"`
}

func (d CreateVehicleDto) Validate() error {
	return validateJolly(d.Jolly, d.JollyDuration)
}

type UpdateVehicleDto struct {
	LicensePlate  *string `json:"license_plate"`
	Brand         *string `json:"brand"`
	Model         *string `json:"model"`
	Year          *int    `json:"year"`
	Color         *string `json:"color"`
	FuelType      *string `json:"fuel_type"`
	Mileage       *int    `json:"mileage"`
	Notes         *string `json:"notes"`
	Jolly         *bool   `json:"jolly"`
	JollyDuration *int    `json:"jolly_duration"`
}

func (d UpdateVehicleDto) Validate(current *model.Vehicle) error {
	jolly := current.Jolly
	if d.Jolly != nil {
		jolly = *d.Jolly
	}
	duration := current.JollyDuration
	if d.JollyDuration != nil {
		duration = *d.JollyDuration
	}
	return validateJolly(jolly, duration)
}

func validateJolly(jolly bool, duration int) error {
	if !jolly && duration != 0 {
		return errors.New("jolly duration can only be set when jolly is true")
	}
	if jolly && duration <= 0 {
		return errors.New("jolly duration must be greater than 0 when jolly is true")
	}
	return nil
}

type AssignVehicleDto struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type ListVehiclesDto struct {
	Status       model.VehicleStatus
	AssignedToID *uuid.UUID
	Search       string
	Limit        int
	Offset       int
}
