package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleDriver Role = "driver"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey"       json:"id"`
	CreatedAt time.Time      `                                  json:"created_at"`
	UpdatedAt time.Time      `                                  json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                      json:"-"`

	FirstName string `gorm:"not null"                   json:"first_name"`
	LastName  string `gorm:"not null"                   json:"last_name"`
	Email     string `gorm:"uniqueIndex;not null"       json:"email"`
	Password  string `gorm:"not null"                   json:"-"`
	Role      Role   `gorm:"type:text;default:'driver'" json:"role"`
	AssignedVehicles []Vehicle `gorm:"foreignKey:AssignedToID"    json:"assigned_vehicles,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
