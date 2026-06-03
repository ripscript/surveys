package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to string, subject string, body string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	message := []byte("Subject: " + subject + "\r\n" +
		"\r\n" + body)
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func SendEmailHtml(to string, subject string, body string, isHTML bool) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	header := "MIME-Version: 1.0\r\n"
	if isHTML {
		header += "Content-Type: text/html; charset=UTF-8\r\n"
	} else {
		header += "Content-Type: text/plain; charset=UTF-8\r\n"
	}

	message := []byte("Subject: " + subject + "\r\n" + header + "\r\n" + body)
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
