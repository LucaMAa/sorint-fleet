package service

import (
	"errors"
	"log"

	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/mailer"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/ws"

	"github.com/google/uuid"
)

type UserService interface {
	List(filters dto.ListUsersDto) ([]model.User, int64, error)
	GetByID(id uuid.UUID) (*model.User, error)
	UpdateRole(id uuid.UUID, role string) (*model.User, error)
	ListPending() ([]model.User, error)
	Approve(id uuid.UUID) (*model.User, error)
	Reject(id uuid.UUID) (*model.User, error)
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

func (s *userService) List(filters dto.ListUsersDto) ([]model.User, int64, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = 10
	}
	return s.userRepo.FindAll(repository.UserFilters{
		Search: filters.Search,
		Limit:  limit,
		Offset: filters.Offset,
	})
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

func (s *userService) ListPending() ([]model.User, error) {
	return s.userRepo.FindByStatus(model.StatusPending)
}

func (s *userService) Approve(id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.Status != model.StatusPending {
		return nil, errors.New("user is not pending")
	}

	user.Status = model.StatusApproved
	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}

	ws.Global.Broadcast(ws.EventUserApproved, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})

	go func() {
		if err := mailer.SendApprovalEmail(user.Email, user.FirstName); err != nil {
			log.Printf("⚠️  Email approvazione non inviata a %s: %v", user.Email, err)
		}
	}()

	return user, nil
}

func (s *userService) Reject(id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.Status != model.StatusPending {
		return nil, errors.New("user is not pending")
	}

	user.Status = model.StatusRejected
	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}

	ws.Global.Broadcast(ws.EventUserRejected, map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})

	return user, nil
}
