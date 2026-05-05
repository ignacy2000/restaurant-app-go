package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer struct {
	from     string
	password string
}

func New(from, password string) *Mailer {
	return &Mailer{from: from, password: password}
}

func (m *Mailer) SendPasswordReset(to, resetURL string) error {
	auth := smtp.PlainAuth("", m.from, m.password, "smtp.gmail.com")
	subject := "Subject: Reset your password\r\n"
	headers := "MIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n"
	body := fmt.Sprintf("Click the link below to reset your password (valid for 1 hour):\r\n\r\n%s\r\n\r\nIf you didn't request a password reset, you can ignore this email.", resetURL)
	msg := []byte("From: " + m.from + "\r\nTo: " + to + "\r\n" + subject + headers + "\r\n" + body)

	if err := smtp.SendMail("smtp.gmail.com:587", auth, m.from, []string{to}, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}
