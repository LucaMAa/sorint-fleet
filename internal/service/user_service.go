package service

import (
	"context"
	"errors"
	"log"

	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/mailer"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/session"
	"sorint-fleet/internal/ws"

	"github.com/google/uuid"
)

type UpdateRoleInput struct {
	Role string `json:"role" binding:"required,oneof=user admin"`
}

type UserService interface {
	List(filters dto.ListUsersParams) ([]model.User, int64, error)
	GetByID(id uuid.UUID) (*model.User, error)
	UpdateRole(id uuid.UUID, role string) (*model.User, error)
	ListPending() ([]model.User, error)
	Approve(id uuid.UUID) (*model.User, error)
	Reject(ctx context.Context, id uuid.UUID) (*model.User, error)
	Enable(id uuid.UUID) (*model.User, error)
	Disable(ctx context.Context, id uuid.UUID) (*model.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	sessions *session.Store
}

func NewUserService(userRepo repository.UserRepository, sessions *session.Store) UserService {
	return &userService{userRepo: userRepo, sessions: sessions}
}

func (s *userService) List(filters dto.ListUsersParams) ([]model.User, int64, error) {
	limit := filters.Limit
	if limit <= 0 {
		limit = 10
	}
	return s.userRepo.FindAll(dto.ListUsersParams{
		PageParams: dto.PageParams{Limit: limit, Offset: filters.Offset},
		Search:     filters.Search,
		Enabled:    filters.Enabled,
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
		"id": user.ID, "email": user.Email,
	})
	go func() {
		if err := mailer.SendApprovalEmail(user.Email, user.FirstName); err != nil {
			log.Printf("⚠️  Email approvazione non inviata a %s: %v", user.Email, err)
		}
	}()
	return user, nil
}

func (s *userService) Reject(ctx context.Context, id uuid.UUID) (*model.User, error) {
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
	if err := s.sessions.DeleteAllForUser(ctx, id); err != nil {
		log.Printf("⚠️  session invalidation on reject for %s: %v", id, err)
	}
	ws.Global.Broadcast(ws.EventUserRejected, map[string]interface{}{
		"id": user.ID, "email": user.Email,
	})
	return user, nil
}

func (s *userService) Enable(id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.Status != model.StatusDisabled {
		return nil, errors.New("user is not disabled")
	}
	user.Status = model.StatusApproved
	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Disable(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.Status == model.StatusDisabled {
		return nil, errors.New("user is already disabled")
	}
	user.Status = model.StatusDisabled
	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}
	if err := s.sessions.DeleteAllForUser(ctx, id); err != nil {
		log.Printf("⚠️  session invalidation on disable for %s: %v", id, err)
	}
	return user, nil
}
