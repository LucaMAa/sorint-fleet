package repository

import (
	"errors"
	"sorint-fleet/internal/config"
	"sorint-fleet/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleAssignmentRepository interface {
	Create(a *model.VehicleAssignment) error
	CloseActive(vehicleID uuid.UUID) error
	FindByVehicle(vehicleID uuid.UUID) ([]model.VehicleAssignment, error)
	FindByUser(userID uuid.UUID) ([]model.VehicleAssignment, error)
	FindActive(vehicleID uuid.UUID) (*model.VehicleAssignment, error)
}

type vehicleAssignmentRepository struct {
	db *gorm.DB
}

func NewVehicleAssignmentRepository() VehicleAssignmentRepository {
	return &vehicleAssignmentRepository{db: config.DB}
}

func (r *vehicleAssignmentRepository) Create(a *model.VehicleAssignment) error {
	return r.db.Create(a).Error
}

func (r *vehicleAssignmentRepository) CloseActive(vehicleID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&model.VehicleAssignment{}).
		Where("vehicle_id = ? AND ended_at IS NULL", vehicleID).
		Update("ended_at", now).Error
}

func (r *vehicleAssignmentRepository) FindByVehicle(vehicleID uuid.UUID) ([]model.VehicleAssignment, error) {
	var list []model.VehicleAssignment
	err := r.db.Preload("User").
		Where("vehicle_id = ?", vehicleID).
		Order("started_at DESC").
		Find(&list).Error
	return list, err
}

func (r *vehicleAssignmentRepository) FindByUser(userID uuid.UUID) ([]model.VehicleAssignment, error) {
	var list []model.VehicleAssignment
	err := r.db.Preload("Vehicle").
		Where("user_id = ?", userID).
		Order("started_at DESC").
		Find(&list).Error
	return list, err
}

func (r *vehicleAssignmentRepository) FindActive(vehicleID uuid.UUID) (*model.VehicleAssignment, error) {
	var a model.VehicleAssignment
	err := r.db.Preload("User").
		Where("vehicle_id = ? AND ended_at IS NULL", vehicleID).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &a, err
}
