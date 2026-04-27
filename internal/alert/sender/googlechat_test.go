package sender_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert/sender"
)

func TestGoogleChatSender_PostsFormattedMessage(t *testing.T) {
	var received map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	s := sender.NewGoogleChatSenderWithURL(srv.URL)

	alert := sender.Alert{
		Level:   "critical",
		LeaseID: "lease/abc123",
		Path:    "secret/db",
		TTL:     30 * time.Minute,
	}

	if err := s.Send(alert); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["text"] == "" {
		t.Fatal("expected non-empty text field")
	}
	if !strings.Contains(received["text"], "critical") {
		t.Errorf("expected level in message, got: %s", received["text"])
	}
	if !strings.Contains(received["text"], "secret/db") {
		t.Errorf("expected path in message, got: %s", received["text"])
	}
}

func TestGoogleChatSender_NonSuccessStatusReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := sender.NewGoogleChatSenderWithURL(srv.URL)

	alert := sender.Alert{
		Level:   "warning",
		LeaseID: "lease/xyz",
		Path:    "secret/api",
		TTL:     10 * time.Minute,
	}

	err := s.Send(alert)
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected status code in error, got: %v", err)
	}
}
