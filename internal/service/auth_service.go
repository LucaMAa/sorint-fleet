package service

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/mailer"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/ws"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(input dto.RegisterDto) error
	Login(input dto.LoginDto) (*dto.AuthResponseDto, error)
	Refresh(refreshToken string) (*dto.AuthResponseDto, error)
	Logout(refreshToken string) error
	GoogleLogin(token string) (*dto.AuthResponseDto, error)
	ChangePassword(userID uuid.UUID, input dto.ChangePasswordDto) error
	RequestPasswordReset(email string) error
	ResetPassword(token, newPassword string) error
}

type authService struct {
	userRepo    repository.UserRepository
	refreshRepo repository.RefreshTokenRepository
	resetRepo   repository.PasswordResetRepository
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	resetRepo repository.PasswordResetRepository,
) AuthService {
	return &authService{userRepo: userRepo, refreshRepo: refreshRepo, resetRepo: resetRepo}
}

func (s *authService) Register(input dto.RegisterDto) error {
	exists, err := s.userRepo.ExistsByEmail(input.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Password:  string(hash),
		Role:      model.RoleUser,
		Status:    model.StatusPending,
	}

	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	ws.Global.Broadcast(ws.EventNewPendingUser, map[string]interface{}{
		"id":         user.ID,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"created_at": user.CreatedAt,
	})

	return nil
}

func (s *authService) Login(input dto.LoginDto) (*dto.AuthResponseDto, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("not valid credentials")
	}

	if user.Status == model.StatusPending {
		return nil, errors.New("account_pending")
	}
	if user.Status == model.StatusRejected {
		return nil, errors.New("account_rejected")
	}

	if user.Status == model.StatusDisabled {
		return nil, errors.New("account_disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, errors.New("not valid credentials")
	}

	token, err := config.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	refresh := uuid.NewString()
	s.refreshRepo.Create(&model.RefreshToken{
		UserID:    user.ID,
		Token:     refresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	return &dto.AuthResponseDto{
		Token:              token,
		RefreshToken:       refresh,
		User:               user,
		MustChangePassword: user.MustChangePassword,
	}, nil
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

	newRefresh := uuid.NewString()
	s.refreshRepo.Create(&model.RefreshToken{
		UserID:    user.ID,
		Token:     newRefresh,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})

	token, _ := config.GenerateToken(user.ID, string(user.Role))

	return &dto.AuthResponseDto{
		Token:              token,
		RefreshToken:       newRefresh,
		User:               user,
		MustChangePassword: user.MustChangePassword,
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
	firstName, lastName := "", ""
	if v, ok := payload.Claims["given_name"].(string); ok {
		firstName = v
	}
	if v, ok := payload.Claims["family_name"].(string); ok {
		lastName = v
	}
	if firstName == "" {
		if v, ok := payload.Claims["name"].(string); ok {
			firstName = v
		}
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
			Status:    model.StatusPending,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, err
		}
		ws.Global.Broadcast(ws.EventNewPendingUser, map[string]interface{}{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"created_at": user.CreatedAt,
		})
		return nil, errors.New("account_pending")
	}

	if user.Status == model.StatusPending {
		return nil, errors.New("account_pending")
	}
	if user.Status == model.StatusRejected {
		return nil, errors.New("account_rejected")
	}

	if user.Status == model.StatusDisabled {
		return nil, errors.New("account_disabled")
	}

	token, err := config.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponseDto{Token: token, User: user}, nil
}

func (s *authService) ChangePassword(userID uuid.UUID, input dto.ChangePasswordDto) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	if !user.MustChangePassword {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword)); err != nil {
			return errors.New("password attuale non corretta")
		}
	}

	if len(input.NewPassword) < 8 {
		return errors.New("la password deve essere di almeno 8 caratteri")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)
	user.MustChangePassword = false
	return s.userRepo.Save(user)
}

func (s *authService) RequestPasswordReset(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return err
	}
	if user == nil || user.Password == "" {
		return nil
	}

	s.resetRepo.DeleteByUserID(user.ID.String())

	token := uuid.NewString()
	s.resetRepo.Create(&model.PasswordReset{
		UserID:    user.ID.String(),
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	})

	go func() {
		if err := mailer.SendResetPasswordEmail(user.Email, user.FirstName, token); err != nil {
			log.Printf("⚠️  Reset email non inviata a %s: %v", user.Email, err)
		}
	}()

	return nil
}

func (s *authService) ResetPassword(token, newPassword string) error {
	pr, err := s.resetRepo.FindByToken(token)
	if err != nil || pr == nil {
		return errors.New("token non valido o scaduto")
	}

	uid, err := uuid.Parse(pr.UserID)
	if err != nil {
		return errors.New("token non valido")
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil || user == nil {
		return errors.New("utente non trovato")
	}

	if len(newPassword) < 8 {
		return errors.New("la password deve essere di almeno 8 caratteri")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hash)
	user.MustChangePassword = false
	if err := s.userRepo.Save(user); err != nil {
		return err
	}

	s.resetRepo.DeleteByUserID(pr.UserID)
	return nil
}
