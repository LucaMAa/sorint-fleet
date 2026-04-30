package service

import (
	"errors"

	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"

	"github.com/google/uuid"
)

type UserService interface {
	List() ([]model.User, error)
	GetByID(id uuid.UUID) (*model.User, error)
	UpdateRole(id uuid.UUID, role string) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

type UpdateRoleInput struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

func (s *userService) List() ([]model.User, error) {
	return s.userRepo.FindAll()
}

func (s *userService) GetByID(id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *userService) UpdateRole(id uuid.UUID, role string) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	if role != string(model.RoleAdmin) && role != string(model.RoleUser) {
		return nil, errors.New("invalid role")
	}

	user.Role = model.Role(role)
	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}

	return user, nil
}
