package repository

import (
	"sorint-fleet/internal/config"
	"sorint-fleet/internal/model"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(t *model.RefreshToken) error
	Find(token string) (*model.RefreshToken, error)
	Delete(token string) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository() RefreshTokenRepository {
	return &refreshTokenRepository{db: config.DB}
}

func (r *refreshTokenRepository) Create(t *model.RefreshToken) error {
	return r.db.Create(t).Error
}

func (r *refreshTokenRepository) Find(token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}
