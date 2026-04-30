package repository

import (
	"errors"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleRepository interface {
	Create(vehicle *model.Vehicle) error
	FindAll(filters VehicleFilters) ([]model.Vehicle, error)
	FindByID(id uuid.UUID) (*model.Vehicle, error)
	FindByLicensePlate(plate string) (*model.Vehicle, error)
	Update(vehicle *model.Vehicle) error
	Delete(id uuid.UUID) error
	FindAllBrands() ([]model.Brand, error)
	FindAllModelsByBrand(brandName string) ([]model.Model, error)
}

type VehicleFilters struct {
	Status       model.VehicleStatus
	AssignedToID *uuid.UUID
}

type vehicleRepository struct {
	db *gorm.DB
}

func NewVehicleRepository() VehicleRepository {
	return &vehicleRepository{db: config.DB}
}

func (r *vehicleRepository) Create(vehicle *model.Vehicle) error {
	return r.db.Create(vehicle).Error
}

func (r *vehicleRepository) FindAll(filters VehicleFilters) ([]model.Vehicle, error) {
	var vehicles []model.Vehicle
	q := r.db.Preload("AssignedTo")

	if filters.Status != "" {
		q = q.Where("status = ?", filters.Status)
	}
	if filters.AssignedToID != nil {
		q = q.Where("assigned_to_id = ?", *filters.AssignedToID)
	}

	err := q.Find(&vehicles).Error
	return vehicles, err
}

func (r *vehicleRepository) FindByID(id uuid.UUID) (*model.Vehicle, error) {
	var vehicle model.Vehicle
	err := r.db.Preload("AssignedTo").First(&vehicle, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &vehicle, err
}

func (r *vehicleRepository) FindByLicensePlate(plate string) (*model.Vehicle, error) {
	var vehicle model.Vehicle
	err := r.db.Where("license_plate = ?", plate).First(&vehicle).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &vehicle, err
}

func (r *vehicleRepository) Update(vehicle *model.Vehicle) error {
	return r.db.Save(vehicle).Error
}

func (r *vehicleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Vehicle{}, "id = ?", id).Error
}

func (r *vehicleRepository) FindAllBrands() ([]model.Brand, error) {
	var brands []model.Brand
	err := r.db.Find(&brands).Error
	return brands, err
}

func (r *vehicleRepository) FindAllModelsByBrand(brandName string) ([]model.Model, error) {
	var models []model.Model
	err := r.db.
		Table("models AS m").
		Select("m.*, b.id AS brand_id, b.name AS brand_name").
		Joins("LEFT JOIN brands b ON m.brand_id = b.id").
		Find(&models).Error

	return models, err
}
