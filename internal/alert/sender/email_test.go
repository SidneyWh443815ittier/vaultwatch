package sender_test

import (
	"net/smtp"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/alert/sender"
)

func TestEmailSender_SendsFormattedEmail(t *testing.T) {
	var capturedMsg []byte
	var capturedTo []string

	s := sender.NewEmailSenderWithFunc(
		"smtp.example.com", 587, "user", "pass",
		"vault@example.com", []string{"ops@example.com"},
		func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			capturedTo = to
			capturedMsg = msg
			return nil
		},
	)

	a := alert.Alert{
		LeaseID:      "secret/data/db#abc123",
		Level:        alert.Warning,
		Expiry:       time.Now().Add(2 * time.Hour),
		TTLRemaining: 2 * time.Hour,
	}

	if err := s.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedTo) != 1 || capturedTo[0] != "ops@example.com" {
		t.Errorf("unexpected recipients: %v", capturedTo)
	}

	msgStr := string(capturedMsg)
	for _, want := range []string{"VaultWatch", "WARNING", "secret/data/db#abc123"} {
		if !containsStr(msgStr, want) {
			t.Errorf("message missing %q", want)
		}
	}
}

func TestEmailSender_PropagatesError(t *testing.T) {
	s := sender.NewEmailSenderWithFunc(
		"smtp.example.com", 587, "user", "pass",
		"vault@example.com", []string{"ops@example.com"},
		func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
			return fmt.Errorf("smtp unavailable")
		},
	)

	a := alert.Alert{LeaseID: "x", Level: alert.Critical}
	if err := s.Send(a); err == nil {
		t.Fatal("expected error, got nil")
	}
}
