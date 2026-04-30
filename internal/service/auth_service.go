package service

import (
	"context"
	"errors"
	"os"
	"time"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(input dto.RegisterDto) (*dto.AuthResponseDto, error)
	Login(input dto.LoginDto) (*dto.AuthResponseDto, error)
	Refresh(refreshToken string) (*dto.AuthResponseDto, error)
	Logout(refreshToken string) error
	GoogleLogin(token string) (*dto.AuthResponseDto, error)
}

type authService struct {
	userRepo    repository.UserRepository
	refreshRepo repository.RefreshTokenRepository
}

func NewAuthService(userRepo repository.UserRepository, refreshRepo repository.RefreshTokenRepository) AuthService {
	return &authService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
	}
}

func (s *authService) Register(input dto.RegisterDto) (*dto.AuthResponseDto, error) {
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

	role := model.RoleUser
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

	refresh := generateRefreshToken()

	s.refreshRepo.Create(&model.RefreshToken{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	return &dto.AuthResponseDto{
		Token:        token,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

func (s *authService) Login(input dto.LoginDto) (*dto.AuthResponseDto, error) {
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

	refresh := generateRefreshToken()

	s.refreshRepo.Create(&model.RefreshToken{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	return &dto.AuthResponseDto{
		Token:        token,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

func generateRefreshToken() string {
	return uuid.NewString()
}

func (s *authService) Refresh(refreshToken string) (*dto.AuthResponseDto, error) {
	rt, err := s.refreshRepo.Find(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if time.Now().After(rt.ExpiresAt) {
		s.refreshRepo.Delete(refreshToken)
		return nil, errors.New("refresh token expired")
	}

	user, err := s.userRepo.FindByID(rt.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	s.refreshRepo.Delete(refreshToken)

	newRefresh := generateRefreshToken()

	s.refreshRepo.Create(&model.RefreshToken{
		UserID:    user.ID,
		Token:     newRefresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	token, _ := config.GenerateToken(user.ID, string(user.Role))

	return &dto.AuthResponseDto{
		Token:        token,
		RefreshToken: newRefresh,
		User:         user,
	}, nil
}

func (s *authService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("missing refresh token")
	}

	return s.refreshRepo.Delete(refreshToken)
}

func (s *authService) GoogleLogin(googleToken string) (*dto.AuthResponseDto, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")

	payload, err := idtoken.Validate(context.Background(), googleToken, clientID)
	if err != nil {
		return nil, errors.New("invalid google token")
	}

	email := payload.Claims["email"].(string)
	firstName := ""
	lastName := ""

	if name, ok := payload.Claims["name"].(string); ok {
		firstName = name
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user = &model.User{
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
			Role:      model.RoleUser,
		}

		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
	}

	token, err := config.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponseDto{
		Token: token,
		User:  user,
	}, nil
}
