package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/mailer"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/session"
	"sorint-fleet/internal/ws"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)
const cookieMaxAge = int(24 * time.Hour / time.Second)

type AuthService interface {
	Login(ctx context.Context, input dto.LoginDto, w http.ResponseWriter) (*dto.AuthResponseDto, error)
	Logout(ctx context.Context, sessionID string, w http.ResponseWriter) error
	GoogleLogin(ctx context.Context, token string, w http.ResponseWriter) (*dto.AuthResponseDto, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, input dto.ChangePasswordDto) error
	RequestPasswordReset(email string) error
	ResetPassword(token, newPassword string) error
	Register(input dto.RegisterDto) error
}

type authService struct {
	userRepo  repository.UserRepository
	resetRepo repository.PasswordResetRepository
	sessions  *session.Store
}

func NewAuthService(
	userRepo repository.UserRepository,
	resetRepo repository.PasswordResetRepository,
	sessions *session.Store,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		resetRepo: resetRepo,
		sessions:  sessions,
	}
}

func (s *authService) createSession(ctx context.Context, user *model.User, w http.ResponseWriter) error {
	sessID, err := s.sessions.Create(ctx, session.Data{
		UserID: user.ID,
		Role:   string(user.Role),
	})
	if err != nil {
		return err
	}
	if err := s.sessions.TrackUserSession(ctx, user.ID, sessID); err != nil {
		log.Printf("⚠️  TrackUserSession: %v", err)
	}
	setSessionCookie(w, sessID)
	return nil
}

func setSessionCookie(w http.ResponseWriter, sessionID string) {
	secure := os.Getenv("COOKIE_SECURE") == "true"
	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     session.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
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


func (s *authService) Login(ctx context.Context, input dto.LoginDto, w http.ResponseWriter) (*dto.AuthResponseDto, error) {
	user, err := s.userRepo.FindByEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("not valid credentials")
	}

	switch user.Status {
	case model.StatusPending:
		return nil, errors.New("account_pending")
	case model.StatusRejected:
		return nil, errors.New("account_rejected")
	case model.StatusDisabled:
		return nil, errors.New("account_disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, errors.New("not valid credentials")
	}

	if err := s.createSession(ctx, user, w); err != nil {
		return nil, err
	}

	return &dto.AuthResponseDto{
		User:               user,
		MustChangePassword: user.MustChangePassword,
	}, nil
}


func (s *authService) Logout(ctx context.Context, sessionID string, w http.ResponseWriter) error {
	clearSessionCookie(w)
	if sessionID == "" {
		return nil
	}
	return s.sessions.Delete(ctx, sessionID)
}


func (s *authService) GoogleLogin(ctx context.Context, googleToken string, w http.ResponseWriter) (*dto.AuthResponseDto, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")

	payload, err := idtoken.Validate(ctx, googleToken, clientID)
	if err != nil {
		return nil, errors.New("invalid google token")
	}

	email, _ := payload.Claims["email"].(string)
	firstName, _ := payload.Claims["given_name"].(string)
	lastName, _ := payload.Claims["family_name"].(string)
	if firstName == "" {
		firstName, _ = payload.Claims["name"].(string)
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

	switch user.Status {
	case model.StatusPending:
		return nil, errors.New("account_pending")
	case model.StatusRejected:
		return nil, errors.New("account_rejected")
	case model.StatusDisabled:
		return nil, errors.New("account_disabled")
	}

	if err := s.createSession(ctx, user, w); err != nil {
		return nil, err
	}

	return &dto.AuthResponseDto{
		User:               user,
		MustChangePassword: user.MustChangePassword,
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uuid.UUID, input dto.ChangePasswordDto) error {
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
