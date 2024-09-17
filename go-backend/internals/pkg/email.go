package pkg

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func SendConfirmationEmail(to, username string) error {
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, smtpServer)

	subject := "DBA Login Password Reset Successful"
	body := fmt.Sprintf("Hello %s,\n\nYour password has been successfully reset.\n\nBest regards,\nYour Company", username)
	message := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))

	addr := fmt.Sprintf("%s:%s", smtpServer, smtpPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, message)
	if err != nil {
		return err
	}

	log.Printf("Confirmation email sent to %s", to)
	return nil
}