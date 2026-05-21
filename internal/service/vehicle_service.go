package service

import (
	"errors"
	"io"
	"time"

	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/validator"
	"sorint-fleet/pkg"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/google/uuid"
)

type VehicleService interface {
	Create(input dto.CreateVehicleDto) (*model.Vehicle, error)
	Update(id uuid.UUID, input dto.UpdateVehicleDto) (*model.Vehicle, error)
	List(filters dto.ListVehiclesDto) ([]model.Vehicle, int64, error)
	GetByID(id uuid.UUID) (*model.Vehicle, error)
	Assign(vehicleID uuid.UUID, input dto.AssignVehicleDto) (*model.Vehicle, error)
	Unassign(vehicleID uuid.UUID) (*model.Vehicle, error)
	Delete(id uuid.UUID) error
	GetBrands() ([]model.Brand, error)
	GetModelsByBrand(brandName string) ([]model.Model, error)
	ImportFromExcel(file io.Reader) (int, error)
}

type vehicleService struct {
	vehicleRepo    repository.VehicleRepository
	userRepo       repository.UserRepository
	assignmentRepo repository.VehicleAssignmentRepository
}

func NewVehicleService(
	vehicleRepo repository.VehicleRepository,
	userRepo repository.UserRepository,
	assignmentRepo repository.VehicleAssignmentRepository,
) VehicleService {
	return &vehicleService{vehicleRepo, userRepo, assignmentRepo}
}

func (s *vehicleService) Create(input dto.CreateVehicleDto) (*model.Vehicle, error) {
	if err := validator.Validate(input); err != nil {
		return nil, err
	}

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
		LicensePlate:  input.LicensePlate,
		Brand:         input.Brand,
		Model:         input.Model,
		Year:          input.Year,
		Color:         input.Color,
		FuelType:      fuelType,
		Mileage:       input.Mileage,
		Notes:         input.Notes,
		Status:        model.StatusAvailable,
		Jolly:         input.Jolly,
		JollyDuration: input.JollyDuration,
	}

	if err := s.vehicleRepo.Create(vehicle); err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *vehicleService) List(filters dto.ListVehiclesDto) ([]model.Vehicle, int64, error) {
	return s.vehicleRepo.FindAll(repository.VehicleFilters{
		Status:       filters.Status,
		AssignedToID: filters.AssignedToID,
		Search:       filters.Search,
		Limit:        filters.Limit,
		Offset:       filters.Offset,
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

func (s *vehicleService) Assign(vehicleID uuid.UUID, input dto.AssignVehicleDto) (*model.Vehicle, error) {
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

	assignment := &model.VehicleAssignment{
		VehicleID: vehicleID,
		UserID:    input.UserID,
		StartedAt: now,
	}
	if err := s.assignmentRepo.Create(assignment); err != nil {
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
	vehicle.AssignedTo = nil
	vehicle.AssignedAt = nil

	if err := s.vehicleRepo.Update(vehicle); err != nil {
		return nil, err
	}
	if err := s.assignmentRepo.CloseActive(vehicleID); err != nil {
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

func (s *vehicleService) GetBrands() ([]model.Brand, error) {
	brands, err := s.vehicleRepo.FindAllBrands()
	if err != nil {
		return []model.Brand{}, err
	}
	return brands, err
}

func (s *vehicleService) GetModelsByBrand(brandName string) ([]model.Model, error) {
	models, err := s.vehicleRepo.FindAllModelsByBrand(brandName)
	if err != nil {
		return []model.Model{}, err
	}
	return models, err
}

func (s *vehicleService) Update(id uuid.UUID, input dto.UpdateVehicleDto) (*model.Vehicle, error) {
	if err := validator.Validate(input); err != nil {
		return nil, err
	}

	vehicle, err := s.vehicleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, errors.New("vehicle not found")
	}

	if input.LicensePlate != nil {
		vehicle.LicensePlate = *input.LicensePlate
	}
	if input.Brand != nil {
		vehicle.Brand = *input.Brand
	}
	if input.Model != nil {
		vehicle.Model = *input.Model
	}
	if input.Year != nil {
		vehicle.Year = *input.Year
	}
	if input.Color != nil {
		vehicle.Color = *input.Color
	}
	if input.FuelType != nil {
		vehicle.FuelType = *input.FuelType
	}
	if input.Mileage != nil {
		vehicle.Mileage = *input.Mileage
	}
	if input.Notes != nil {
		vehicle.Notes = *input.Notes
	}
	if input.Jolly != nil {
		vehicle.Jolly = *input.Jolly
	}
	if input.JollyDuration != nil {
		vehicle.JollyDuration = *input.JollyDuration
	}

	if err := s.vehicleRepo.Update(vehicle); err != nil {
		return nil, err
	}

	return s.vehicleRepo.FindByID(id)
}

func (s *vehicleService) ImportFromExcel(file io.Reader) (int, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		return 0, err
	}

	sheets := f.GetSheetMap()
	if len(sheets) == 0 {
		return 0, errors.New("no sheets found")
	}

	var sheetName string
	for _, name := range sheets {
		sheetName = name
		break
	}

	reader, err := f.Rows(sheetName)
	if err != nil {
		return 0, err
	}

	vehicles := make([]*model.Vehicle, 0)
	first := true

	for reader.Next() {
		row := reader.Columns()
		if first {
			first = false
			continue
		}
		if len(row) < 4 {
			continue
		}
		vehicles = append(vehicles, &model.Vehicle{
			LicensePlate: row[0],
			Brand:        row[1],
			Model:        row[2],
			Year:         pkg.ParseInt(row[3]),
			Color:        pkg.GetOr(row, 4),
			FuelType:     pkg.GetOr(row, 5),
			Mileage:      pkg.ParseInt(pkg.GetOr(row, 6)),
			Notes:        pkg.GetOr(row, 7),
			Status:       model.StatusAvailable,
		})
	}

	if len(vehicles) == 0 {
		return 0, errors.New("no valid rows")
	}

	if err := s.vehicleRepo.CreateBatch(vehicles); err != nil {
		return 0, err
	}

	return len(vehicles), nil
}
