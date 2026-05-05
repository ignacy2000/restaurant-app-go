package mailer

import (
	"fmt"
	"net/smtp"
	"strings"
)

type OrderItem struct {
	Name     string
	Quantity int
}

type Mailer struct {
	from     string
	password string
}

func New(from, password string) *Mailer {
	return &Mailer{from: from, password: password}
}

func (m *Mailer) send(to, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", m.from, m.password, "smtp.gmail.com")
	headers := strings.Join([]string{
		"From: " + m.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}, "\r\n")
	msg := []byte(headers + "\r\n\r\n" + htmlBody)
	return smtp.SendMail("smtp.gmail.com:587", auth, m.from, []string{to}, msg)
}

func (m *Mailer) SendPasswordReset(to, resetURL string) error {
	body := fmt.Sprintf(`<p>Click the link below to reset your password (valid for 1 hour):</p><p><a href="%s">%s</a></p><p>If you didn't request a password reset, you can ignore this email.</p>`, resetURL, resetURL)
	if err := m.send(to, "Reset your password", body); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

func (m *Mailer) SendOrderConfirmation(to, restaurantName string, tableNumber int, items []OrderItem, notes, confirmURL string) error {
	var rows strings.Builder
	for _, it := range items {
		rows.WriteString(fmt.Sprintf(`<tr><td style="padding:4px 8px;">×%d</td><td style="padding:4px 8px;">%s</td></tr>`, it.Quantity, it.Name))
	}

	notesSection := ""
	if notes != "" {
		notesSection = fmt.Sprintf(`<p style="color:#555;"><strong>Uwagi:</strong> %s</p>`, notes)
	}

	restaurantLabel := restaurantName
	if restaurantLabel == "" {
		restaurantLabel = "restauracji"
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:sans-serif;max-width:560px;margin:0 auto;padding:24px;color:#111;">
<h2 style="color:#2563eb;margin-bottom:4px;">Potwierdź swoje zamówienie</h2>
<p style="color:#555;margin-top:0;">Dziękujemy za zamówienie w <strong>%s</strong>!</p>
<table style="border-collapse:collapse;margin-bottom:16px;">
  <tr><td style="padding:4px 8px;color:#888;">Stolik</td><td style="padding:4px 8px;font-weight:bold;">#%d</td></tr>
</table>
<h3 style="margin-bottom:8px;">Zamówione pozycje</h3>
<table style="border-collapse:collapse;width:100%%;background:#f9fafb;border-radius:8px;">%s</table>
%s
<div style="margin:28px 0;text-align:center;">
  <a href="%s" style="display:inline-block;background:#2563eb;color:#fff;font-weight:bold;font-size:16px;padding:14px 32px;border-radius:10px;text-decoration:none;">
    Potwierdź zamówienie
  </a>
</div>
<p style="font-size:12px;color:#aaa;">Link wygasa po 1 godzinie. Jeśli to nie Ty złożyłeś/aś zamówienie, zignoruj tę wiadomość.</p>
</body></html>`,
		restaurantLabel, tableNumber, rows.String(), notesSection, confirmURL,
	)

	if err := m.send(to, fmt.Sprintf("Potwierdzenie zamówienia — Stolik #%d", tableNumber), body); err != nil {
		return fmt.Errorf("send order confirmation: %w", err)
	}
	return nil
}
