package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type FormSubmissionStatus string

const (
	SubmissionStatusPending  FormSubmissionStatus = "pending"
	SubmissionStatusApproved FormSubmissionStatus = "approved"
	SubmissionStatusRejected FormSubmissionStatus = "rejected"
)

type FormSubmission struct {
	ID             uuid.UUID            `gorm:"type:text;primaryKey" json:"id"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
	DeletedAt      gorm.DeletedAt       `gorm:"index" json:"-"`

	FormTemplateID uuid.UUID            `gorm:"type:text;not null;index" json:"form_template_id"`
	FormTemplate   FormTemplate         `gorm:"foreignKey:FormTemplateID" json:"form_template,omitempty"`

	FirstName      string               `json:"first_name"`
	LastName       string               `json:"last_name"`
	Data           datatypes.JSON       `gorm:"type:jsonb;not null" json:"data"`
	Status         FormSubmissionStatus `gorm:"type:text;default:'pending'" json:"status"`
	AdminNotes     string               `json:"admin_notes,omitempty"`
	SubmittedByID  *uuid.UUID           `gorm:"type:text;index" json:"submitted_by_id,omitempty"`
}

func (s *FormSubmission) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
