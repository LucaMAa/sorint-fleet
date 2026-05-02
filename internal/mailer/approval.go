package mailer

import "os"

type approvalData struct {
	FirstName string
	AppURL    string
}

func SendApprovalEmail(to, firstName string) error {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:5173"
	}

	body, err := render("approval.html", approvalData{
		FirstName: firstName,
		AppURL:    appURL,
	})
	if err != nil {
		return err
	}

	return Send(to, "Il tuo account Fleet è stato approvato!", body)
}
