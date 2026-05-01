package dto

import (
	"sorint-fleet/internal/model"

	"github.com/google/uuid"
)

type CreateVehicleDto struct {
	LicensePlate string `json:"license_plate" binding:"required"`
	Brand        string `json:"brand"         binding:"required"`
	Model        string `json:"model"         binding:"required"`
	Year         int    `json:"year"          binding:"required,min=1900"`
	Color        string `json:"color"`
	FuelType     string `json:"fuel_type"`
	Mileage      int    `json:"mileage"`
	Notes        string `json:"notes"`
}

type UpdateVehicleDto struct {
	LicensePlate *string `json:"license_plate"`
	Brand        *string `json:"brand"`
	Model        *string `json:"model"`
	Year         *int    `json:"year"`
	Color        *string `json:"color"`
	FuelType     *string `json:"fuel_type"`
	Mileage      *int    `json:"mileage"`
	Notes        *string `json:"notes"`
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

type ListUsersDto struct {
	Search string
	Limit  int
	Offset int
}
