package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(toEmail string, subject string, body string) error {
	// TEMPORARILY HARDCODE THESE TWO LINES:
	from := "lekefrnk@gmail.com"
	password := "sifbgxzjfqgfzsad"

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// The message format requires these headers so Gmail reads it correctly
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", toEmail, subject, body))

	// Authenticate with Google
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{toEmail}, message)
	if err != nil {
		return err
	}
	return nil
}
