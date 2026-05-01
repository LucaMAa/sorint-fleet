package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type UserStatus string

const (
	StatusPending  UserStatus = "pending"
	StatusApproved UserStatus = "approved"
	StatusRejected UserStatus = "rejected"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:text;primaryKey"         json:"id"`
	CreatedAt time.Time      `                                    json:"created_at"`
	UpdatedAt time.Time      `                                    json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                        json:"-"`

	FirstName          string     `gorm:"not null"                     json:"first_name"`
	LastName           string     `gorm:"not null"                     json:"last_name"`
	Email              string     `gorm:"uniqueIndex;not null"         json:"email"`
	Password           string     `gorm:"not null"                     json:"-"`
	Role               Role       `gorm:"type:text;default:'user'"     json:"role"`
	Status             UserStatus `gorm:"type:text;default:'approved'" json:"status"`
	MustChangePassword bool       `gorm:"default:false"                json:"must_change_password"`

	AssignedVehicles []Vehicle `gorm:"foreignKey:AssignedToID" json:"assigned_vehicles,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
