package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// baseDir risolve il path dei template relativamente a questo file
func baseDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "templates")
}

func render(templateName string, data any) (string, error) {
	tmplPath := filepath.Join(baseDir(), templateName)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("template %s: %w", templateName, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func Send(to, subject, htmlBody string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if host == "" || user == "" || pass == "" {
		return nil
	}
	if from == "" {
		from = user
	}

	auth := smtp.PlainAuth("", user, pass, host)

	header := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}, "\r\n")

	msg := []byte(header + "\r\n\r\n" + htmlBody)
	return smtp.SendMail(fmt.Sprintf("%s:%s", host, port), auth, user, []string{to}, msg)
}

func adminEmailFromEnv() string {
	e := os.Getenv("ADMIN_EMAIL")
	if e == "" {
		return "admin@sorint.it"
	}
	return e
}
