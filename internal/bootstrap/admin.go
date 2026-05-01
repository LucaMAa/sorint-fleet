package bootstrap

import (
	"log"
	"os"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/model"

	"golang.org/x/crypto/bcrypt"
)

func Admin() {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" || adminPassword == "" {
		log.Println("ADMIN bootstrap skipped")
		return
	}

	var user model.User
	result := config.DB.Where("email = ?", adminEmail).First(&user)

	if result.Error == nil {
		log.Println("admin already exists")
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	admin := model.User{
		Email:              adminEmail,
		FirstName:          "Admin",
		LastName:           "Fleet",
		Password:           string(hash),
		Role:               model.RoleAdmin,
		Status:             model.StatusApproved,
		MustChangePassword: true,
	}

	config.DB.Create(&admin)

	log.Println("Admin created — must change password on first login")
}
