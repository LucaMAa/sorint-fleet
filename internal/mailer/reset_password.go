package mailer

import (
	"fmt"
	"os"
)

type resetPasswordData struct {
	FirstName string
	ResetURL  string
}

func SendResetPasswordEmail(to, firstName, token string) error {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:5173"
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", appURL, token)

	body, err := render("reset_password.html", resetPasswordData{
		FirstName: firstName,
		ResetURL:  resetURL,
	})
	if err != nil {
		return err
	}

	return Send(to, "Reset password Fleet", body)
}
