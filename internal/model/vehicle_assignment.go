// internal/model/vehicle_assignment.go
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleAssignment struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey"  json:"id"`
	CreatedAt time.Time      `                             json:"created_at"`
	UpdatedAt time.Time      `                             json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                 json:"-"`

	VehicleID uuid.UUID  `gorm:"type:text;not null;index" json:"vehicle_id"`
	UserID    uuid.UUID  `gorm:"type:text;not null;index" json:"user_id"`
	StartedAt time.Time  `gorm:"not null"                 json:"started_at"`
	EndedAt   *time.Time `                                 json:"ended_at,omitempty"`
	Notes     string     `                                 json:"notes"`

	Vehicle *Vehicle `gorm:"foreignKey:VehicleID" json:"vehicle,omitempty"`
	User    *User    `gorm:"foreignKey:UserID"    json:"user,omitempty"`
	NotifiedAt *time.Time `json:"notified_at,omitempty"`
}

func (a *VehicleAssignment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
