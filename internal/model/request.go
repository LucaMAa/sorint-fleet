package model

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Request struct {
    ID        uuid.UUID      `gorm:"type:text;primaryKey" json:"id"`
    UserID    *uuid.UUID     `gorm:"type:text;index" json:"user_id,omitempty"`
    QueryText string         `gorm:"type:text" json:"query_text"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *Request) BeforeCreate(tx *gorm.DB) error {
    if r.ID == uuid.Nil {
        r.ID = uuid.New()
    }
    return nil
}
