package service

import (
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"

	"github.com/google/uuid"
)

type VehicleAssignmentService interface {
	GetByVehicle(vehicleID uuid.UUID) ([]model.VehicleAssignment, error)
	GetByUser(userID uuid.UUID) ([]model.VehicleAssignment, error)
}

type vehicleAssignmentService struct {
	repo repository.VehicleAssignmentRepository
}

func NewVehicleAssignmentService(repo repository.VehicleAssignmentRepository) VehicleAssignmentService {
	return &vehicleAssignmentService{repo}
}

func (s *vehicleAssignmentService) GetByVehicle(id uuid.UUID) ([]model.VehicleAssignment, error) {
	return s.repo.FindByVehicle(id)
}

func (s *vehicleAssignmentService) GetByUser(id uuid.UUID) ([]model.VehicleAssignment, error) {
	return s.repo.FindByUser(id)
}
