package dto

import "sorint-fleet/internal/model"

type RegisterDto struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
	Email     string `json:"email"      binding:"required,email"`
	Password  string `json:"password"   binding:"required,min=8"`
	Role      string `json:"role"`
}

type LoginDto struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponseDto struct {
	Token              string      `json:"token"`
	RefreshToken       string      `json:"refresh_token"`
	User               *model.User `json:"user"`
	MustChangePassword bool        `json:"must_change_password"`
}

type GoogleAuthDto struct {
	Token string `json:"token" binding:"required"`
}

type ChangePasswordDto struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}
