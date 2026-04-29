package service

import (
	"errors"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
	Email     string `json:"email"      binding:"required,email"`
	Password  string `json:"password"   binding:"required,min=8"`
	Role string `json:"role"`
}

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *model.User `json:"user"`
}

type AuthService interface {
	Register(input RegisterInput) (*AuthResponse, error)
	Login(input LoginInput) (*AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Register(input RegisterInput) (*AuthResponse, error) {
	exists, err := s.userRepo.ExistsByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := model.RoleDriver
	if input.Role == string(model.RoleAdmin) {
		role = model.RoleAdmin
	}

	user := &model.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  string(hash),
		Role:      role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := config.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *authService) Login(input LoginInput) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("not valid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, errors.New("not valid credentials")
	}

	token, err := config.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}
