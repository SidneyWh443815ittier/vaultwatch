package sender

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/vaultwatch/internal/alert"
)

// EmailSender sends alert notifications via SMTP.
type EmailSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
	to       []string
}

// NewEmailSender creates a new EmailSender.
func NewEmailSender(host string, port int, user, password, from string, to []string) *EmailSender {
	return &EmailSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
		to:       to,
	}
}

// Send delivers an alert via email.
func (e *EmailSender) Send(a alert.Alert) error {
	addr := fmt.Sprintf("%s:%d", e.host, e.port)
	auth := smtp.PlainAuth("", e.user, e.password, e.host)

	subject := fmt.Sprintf("[VaultWatch] %s - %s", strings.ToUpper(string(a.Level)), a.LeaseID)
	body := fmt.Sprintf(
		"Lease: %s\nStatus: %s\nExpires: %s\nTTL Remaining: %s\n",
		a.LeaseID, a.Level, a.Expiry.Format("2006-01-02 15:04:05 UTC"), a.TTLRemaining,
	)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		e.from, strings.Join(e.to, ", "), subject, body,
	))

	return smtp.SendMail(addr, auth, e.from, e.to, msg)
}
