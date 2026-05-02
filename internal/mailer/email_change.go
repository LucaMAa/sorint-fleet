package mailer

import (
	"fmt"
	"os"
)

type emailChangeData struct {
	FirstName  string
	NewEmail   string
	ConfirmURL string
}

func SendEmailChangeConfirmation(to, firstName, newEmail, token string) error {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:5173"
	}

	confirmURL := fmt.Sprintf("%s/confirm-email?token=%s", appURL, token)

	body, err := render("email_change.html", emailChangeData{
		FirstName:  firstName,
		NewEmail:   newEmail,
		ConfirmURL: confirmURL,
	})
	if err != nil {
		return err
	}

	return Send(to, "Conferma il cambio email — Fleet", body)
}
