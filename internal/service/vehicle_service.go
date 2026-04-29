package service

import (
	"errors"
	"time"

	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"

	"github.com/google/uuid"
)

// DTO

type CreateVehicleInput struct {
	LicensePlate string `json:"license_plate" binding:"required"`
	Brand        string `json:"brand"         binding:"required"`
	Model        string `json:"model"         binding:"required"`
	Year         int    `json:"year"          binding:"required,min=1900"`
	Color        string `json:"color"`
	FuelType     string `json:"fuel_type"`
	Mileage      int    `json:"mileage"`
	Notes        string `json:"notes"`
}

type AssignVehicleInput struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

type ListVehiclesInput struct {
	Status       model.VehicleStatus
	AssignedToID *uuid.UUID
}

type VehicleService interface {
	Create(input CreateVehicleInput) (*model.Vehicle, error)
	List(filters ListVehiclesInput) ([]model.Vehicle, error)
	GetByID(id uuid.UUID) (*model.Vehicle, error)
	Assign(vehicleID uuid.UUID, input AssignVehicleInput) (*model.Vehicle, error)
	Unassign(vehicleID uuid.UUID) (*model.Vehicle, error)
	Delete(id uuid.UUID) error
}

type vehicleService struct {
	vehicleRepo repository.VehicleRepository
	userRepo    repository.UserRepository
}

func NewVehicleService(
	vehicleRepo repository.VehicleRepository,
	userRepo repository.UserRepository,
) VehicleService {
	return &vehicleService{
		vehicleRepo: vehicleRepo,
		userRepo:    userRepo,
	}
}

func (s *vehicleService) Create(input CreateVehicleInput) (*model.Vehicle, error) {
	existing, err := s.vehicleRepo.FindByLicensePlate(input.LicensePlate)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("license plate already exist")
	}

	fuelType := input.FuelType
	if fuelType == "" {
		fuelType = "gas"
	}

	vehicle := &model.Vehicle{
		LicensePlate: input.LicensePlate,
		Brand:        input.Brand,
		Model:        input.Model,
		Year:         input.Year,
		Color:        input.Color,
		FuelType:     fuelType,
		Mileage:      input.Mileage,
		Notes:        input.Notes,
		Status:       model.StatusAvailable,
	}

	if err := s.vehicleRepo.Create(vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) List(filters ListVehiclesInput) ([]model.Vehicle, error) {
	return s.vehicleRepo.FindAll(repository.VehicleFilters{
		Status:       filters.Status,
		AssignedToID: filters.AssignedToID,
	})
}

func (s *vehicleService) GetByID(id uuid.UUID) (*model.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, errors.New("vehicle not found")
	}
	return vehicle, nil
}

func (s *vehicleService) Assign(vehicleID uuid.UUID, input AssignVehicleInput) (*model.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, errors.New("vehicle not found")
	}
	if vehicle.Status == model.StatusAssigned {
		return nil, errors.New("vehicle already assigned")
	}
	if vehicle.Status == model.StatusMaintenance {
		return nil, errors.New("vehicle under maintenance, not assignable")
	}

	user, err := s.userRepo.FindByID(input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	now := time.Now()
	vehicle.Status = model.StatusAssigned
	vehicle.AssignedToID = &input.UserID
	vehicle.AssignedAt = &now

	if err := s.vehicleRepo.Update(vehicle); err != nil {
		return nil, err
	}

	return s.vehicleRepo.FindByID(vehicleID)
}

func (s *vehicleService) Unassign(vehicleID uuid.UUID) (*model.Vehicle, error) {
	vehicle, err := s.vehicleRepo.FindByID(vehicleID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, errors.New("vehicle not found")
	}
	if vehicle.Status != model.StatusAssigned {
		return nil, errors.New("the vehicle is not assigned")
	}

	vehicle.Status = model.StatusAvailable
	vehicle.AssignedToID = nil
	vehicle.AssignedAt = nil

	if err := s.vehicleRepo.Update(vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) Delete(id uuid.UUID) error {
	vehicle, err := s.vehicleRepo.FindByID(id)
	if err != nil {
		return err
	}
	if vehicle == nil {
		return errors.New("vehicle not found")
	}
	return s.vehicleRepo.Delete(id)
}
