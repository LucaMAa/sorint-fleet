package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleStatus string

const (
	StatusAvailable   VehicleStatus = "available"
	StatusAssigned    VehicleStatus = "assigned"
	StatusMaintenance VehicleStatus = "maintenance"
)

type Vehicle struct {
	ID            uuid.UUID      `gorm:"type:text;primaryKey"         json:"id"`
	CreatedAt     time.Time      `                                    json:"created_at"`
	UpdatedAt     time.Time      `                                    json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                        json:"-"`
	LicensePlate  string         `gorm:"uniqueIndex;not null"         json:"license_plate"`
	Brand         string         `gorm:"not null"                     json:"brand"`
	Model         string         `gorm:"not null"                     json:"model"`
	Year          int            `gorm:"not null"                     json:"year"`
	Color         string         `                                    json:"color"`
	FuelType      string         `gorm:"default:'benzina'"            json:"fuel_type"`
	Status        VehicleStatus  `gorm:"type:text;default:'available'" json:"status"`
	AssignedToID  *uuid.UUID     `gorm:"type:text;index"               json:"assigned_to_id,omitempty"`
	AssignedTo    *User          `gorm:"foreignKey:AssignedToID"       json:"assigned_to,omitempty"`
	AssignedAt    *time.Time     `                                     json:"assigned_at,omitempty"`
	Mileage       int            `gorm:"default:0"                    json:"mileage"`
	Notes         string         `                                    json:"notes"`
	Jolly         bool           `                                      json:"jolly"`
	JollyDuration int            `                                      json:"jolly_duration"`
}

func (v *Vehicle) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return v.validate()
}

func (v *Vehicle) BeforeSave(tx *gorm.DB) error {
	return v.validate()
}

func (v *Vehicle) validate() error {
	if !v.Jolly && v.JollyDuration != 0 {
		return errors.New("jolly duration can only be set when jolly is true")
	}
	if v.Jolly && v.JollyDuration <= 0 {
		return errors.New("jolly duration must be greater than 0 when jolly is true")
	}
	return nil
}
