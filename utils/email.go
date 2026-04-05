package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(toEmail string, subject string, body string) error {
	// 1. The credentials used to LOG IN to Brevo (Pulled safely from Render)
	smtpUsername := os.Getenv("SMTP_EMAIL")    // e.g., a73610001@smtp-brevo.com
	smtpPassword := os.Getenv("SMTP_PASSWORD") // Your secret Brevo Key
	smtpHost := os.Getenv("SMTP_HOST")         // smtp-relay.brevo.com
	smtpPort := os.Getenv("SMTP_PORT")         // 2525

	// 2. The actual email address the message is SENT FROM (Must be verified in Brevo!)
	// If you used a different email to sign up for Brevo, change this string!
	senderEmail := "bamidelefelix69@gmail.com"

	// 3. Construct the email message (Notice we added the "From:" header)
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", toEmail, senderEmail, subject, body))

	// 4. Authenticate using the SMTP Username (NOT the sender email)
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)

	// 5. Send the email using the senderEmail for the envelope
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{toEmail}, message)
	if err != nil {
		return err
	}

	return nil
}
