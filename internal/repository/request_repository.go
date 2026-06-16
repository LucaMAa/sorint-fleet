package repository

import (
	"sorint-fleet/internal/config"
	"sorint-fleet/internal/model"

	"gorm.io/gorm"
)

type RequestRepository struct {
	db *gorm.DB
}

func NewRequestRepository() *RequestRepository {
	return &RequestRepository{db: config.DB}
}

func (r *RequestRepository) Create(req *model.Request) error {
	return r.db.Create(req).Error
}
