package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FormTemplate struct {
	ID          uuid.UUID      `gorm:"type:text;primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Slug        string         `gorm:"uniqueIndex;not null" json:"slug"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Active      bool           `gorm:"default:true" json:"active"`
	Fields      datatypes.JSON `gorm:"type:jsonb;not null" json:"fields"`
}

func (f *FormTemplate) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
