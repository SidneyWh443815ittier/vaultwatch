package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestGoogleChatSender_PostsFormattedMessage(t *testing.T) {
	var captured map[string]string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := sender.NewGoogleChatSenderWithURL(ts.URL)
	err := s.Send(alert.Alert{
		Level:   alert.Warning,
		LeaseID: "secret/my-app/token",
		TTL:     2 * time.Hour,
		Message: "Lease expiring soon",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	text, ok := captured["text"]
	if !ok {
		t.Fatal("expected 'text' field in payload")
	}
	if !strings.Contains(text, "secret/my-app/token") {
		t.Errorf("expected lease ID in message, got: %s", text)
	}
	if !strings.Contains(text, "WARNING") {
		t.Errorf("expected level in message, got: %s", text)
	}
	if !strings.Contains(text, "Lease expiring soon") {
		t.Errorf("expected message body in payload, got: %s", text)
	}
}

func TestGoogleChatSender_NonSuccessStatusReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := sender.NewGoogleChatSenderWithURL(ts.URL)
	err := s.Send(alert.Alert{
		Level:   alert.Critical,
		LeaseID: "secret/db/creds",
		TTL:     10 * time.Minute,
		Message: "Lease critical",
	})

	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
